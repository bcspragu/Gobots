package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"zombiezen.com/go/capnproto2"

	"github.com/bcspragu/Gobots/botapi"
	"github.com/boltdb/bolt"
)

var (
	errDatastoreNotImplemented = errors.New("gobots: datastore operation not implemented")
	errUserNotFound            = errors.New("gobots: user not found")
	errAINotFound              = errors.New("gobots: AI not found")
	errGameNotFound            = errors.New("gobots: game not found")
)

type WinType int

const (
	P1Win WinType = iota
	P2Win
	Tie
)

type datastore interface {
	// Users
	createUser(u *userInfo) (id uID, err error)
	loadUser(a accessToken) (*userInfo, error)
	userExists(name string) (bool, error)

	// AIs
	createAI(info *aiInfo, a accessToken) (id aiID, err error)
	listAIsForUser(a accessToken) ([]*aiInfo, error)
	lookupAI(id aiID) (*aiInfo, error)
	loadDirectory() (*directory, error)

	// Games
	startGame(ai1, ai2 aiID, init botapi.InitialBoard) (gameID, error)
	addRound(id gameID, round botapi.Replay_Round) error
	lookupGame(id gameID) (botapi.Replay, error)
	lookupGameInfo(id gameID) (*gameInfo, error)
	finishGame(id gameID, p1, p2 *aiInfo, w WinType) error
}

type dbImpl struct {
	*bolt.DB
}

var (
	AIBucket      = []byte("AI")      // aID -> aiInfo
	AIStatsBucket = []byte("AIStats") // aID -> aiInfo

	GameBucket     = []byte("Games")
	GameInfoBucket = []byte("GameInfo") // Which AI were in the match

	UserBucket       = []byte("Users")       // accessToken -> userInfo
	UserLookupBucket = []byte("UserLookups") // userInfo.Name -> []byte{}
)

func initDB(dbName string) (datastore, error) {
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{AIBucket, AIStatsBucket, GameBucket, GameInfoBucket, UserBucket, UserLookupBucket} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}

		return nil
	})

	return &dbImpl{db}, err
}

type directory struct {
	Usernames map[uID]string
	AIs       map[aiID]*aiInfo
	AIStats   map[aiID]*aiStats
}

type (
	accessToken string
	uID         string
	aiID        string
	gameID      string
)

type userInfo struct {
	ID    uID
	Name  string
	Token accessToken

	AIs []*aiInfo
}

type aiInfo struct {
	ID     aiID
	Name   string
	UserID uID
}

type gameInfo struct {
	AI1 *aiInfo
	AI2 *aiInfo
}

type aiStats struct {
	Wins   int
	Losses int
	Ties   int
}

func (db *dbImpl) createUser(uInfo *userInfo) (id uID, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := writeUser(tx, uInfo); err != nil {
			return err
		}

		b := tx.Bucket(UserLookupBucket)
		log.Printf("CreateUser: Putting empty bytes UserLookupBucket [%s]=[]byte{}", uInfo.Name)
		if err := b.Put([]byte(strings.ToLower(uInfo.Name)), []byte{}); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserLookupBucket: " + err.Error())
		}
		return nil
	})
	return
}

func (db *dbImpl) loadUser(a accessToken) (*userInfo, error) {
	var u *userInfo
	err := db.View(func(tx *bolt.Tx) error {
		v, err := user(tx, a)
		u = v
		return err
	})

	return u, err
}

func (db *dbImpl) userExists(name string) (exists bool, err error) {
	name = strings.ToLower(name)
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserLookupBucket)
		// If the bucket has that key, the user exists
		exists = b.Get([]byte(name)) != nil
		log.Printf("UserExists: Loading UserLookupBucket at [%s], exists is %t", name, exists)
		return nil
	})
	return
}

// AIs
func (db *dbImpl) createAI(info *aiInfo, a accessToken) (id aiID, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		u, err := user(tx, a)
		if err != nil {
			return err
		}

		for _, ai := range u.AIs {
			if ai.Name == info.Name {
				log.Printf("CreateAI: Bot %s already exists, not creating", info.Name)
				return errors.New("Bot already exists")
			}
		}

		newInfo := &aiInfo{
			Name:   info.Name,
			UserID: u.ID,
		}
		id, err = writeAi(tx, newInfo)
		if err != nil {
			return err
		}

		buf := new(bytes.Buffer)
		b := tx.Bucket(UserBucket)
		// Append the AI to the list of AIs for this user and save that
		u.AIs = append(u.AIs, newInfo)
		if err := gob.NewEncoder(buf).Encode(u); err != nil {
			return err
		}
		log.Printf("CreateAI: Putting into UserBucket at [%s]=%v", a, buf.Bytes())
		if err := b.Put([]byte(a), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing updated AIs to UserBucket: " + err.Error())
		}

		// Write an empty AI stats to show we're adding this AI
		return writeAiStats(tx, id, &aiStats{})
	})
	return
}

func (db *dbImpl) listAIsForUser(a accessToken) ([]*aiInfo, error) {
	var info userInfo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		dat := b.Get([]byte(a))
		log.Printf("ListAIsForUser: Loading from UserBucket at [%s]=%v", a, dat)
		if len(dat) == 0 {
			return errUserNotFound
		}

		buf := bytes.NewReader(dat)
		return gob.NewDecoder(buf).Decode(&info)
	})
	log.Printf("ListAIsForUser: Decoding from UserBucket at [%s]=%v", a, info)
	return info.AIs, err
}

func (db *dbImpl) lookupAI(id aiID) (*aiInfo, error) {
	var info *aiInfo
	err := db.View(func(tx *bolt.Tx) error {
		ai, err := ai(tx, id)
		info = ai
		return err
	})
	return info, err
}

func (db *dbImpl) loadDirectory() (*directory, error) {
	dir := &directory{
		Usernames: make(map[uID]string),
		AIs:       make(map[aiID]*aiInfo),
		AIStats:   make(map[aiID]*aiStats),
	}
	err := db.View(func(tx *bolt.Tx) error {
		// Load AIs
		b := tx.Bucket(AIBucket)
		err := b.ForEach(func(k, v []byte) error {
			var a aiInfo
			buf := bytes.NewReader(v)
			dec := gob.NewDecoder(buf)

			if err := dec.Decode(&a); err != nil {
				return err
			}
			dir.AIs[a.ID] = &a
			return nil
		})
		if err != nil {
			return err
		}

		// Load AIStats
		b = tx.Bucket(AIStatsBucket)
		err = b.ForEach(func(k, v []byte) error {
			var a aiStats
			buf := bytes.NewReader(v)
			dec := gob.NewDecoder(buf)

			if err := dec.Decode(&a); err != nil {
				return err
			}
			dir.AIStats[aiID(k)] = &a
			return nil
		})
		if err != nil {
			return err
		}

		// Load usernames
		b = tx.Bucket(UserBucket)
		return b.ForEach(func(k, v []byte) error {
			var u userInfo
			buf := bytes.NewReader(v)
			dec := gob.NewDecoder(buf)

			if err := dec.Decode(&u); err != nil {
				return err
			}
			dir.Usernames[u.ID] = u.Name
			return nil
		})
	})
	return dir, err
}

// Games
func (db *dbImpl) startGame(ai1, ai2 aiID, init botapi.InitialBoard) (gameID, error) {
	var gID gameID
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(GameBucket)

		msg, s, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return err
		}
		r, err := botapi.NewRootReplay(s)
		if err != nil {
			return err
		}
		idNum, err := b.NextSequence()
		if err != nil {
			return err
		}
		gID = gameID(strconv.FormatUint(idNum, 10))
		r.SetGameId(string(gID))
		r.SetInitial(init)

		data, err := msg.Marshal()
		if err != nil {
			return err
		}

		return b.Put([]byte(gID), data)
	})

	return gID, err

}

func (db *dbImpl) addRound(id gameID, round botapi.Replay_Round) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(GameBucket)
		key := []byte(id)
		data := b.Get(key)
		if len(data) == 0 {
			return errGameNotFound
		}
		msg, err := capnp.Unmarshal(copyBytes(data))
		if err != nil {
			return err
		}
		orig, err := botapi.ReadRootReplay(msg)
		if err != nil {
			return err
		}
		newMsg, err := addReplayRound(orig, round)
		if err != nil {
			return err
		}
		newData, err := newMsg.Marshal()
		if err != nil {
			return err
		}
		return b.Put(key, newData)
	})
}

func addReplayRound(orig botapi.Replay, round botapi.Replay_Round) (*capnp.Message, error) {
	newMsg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, err
	}
	newReplay, err := botapi.NewRootReplay(seg)
	if err != nil {
		return nil, err
	}
	gid, err := orig.GameId()
	if err != nil {
		return nil, err
	}
	if err := newReplay.SetGameId(gid); err != nil {
		return nil, err
	}
	initBoard, err := orig.Initial()
	if err != nil {
		return nil, err
	}
	if err := newReplay.SetInitial(initBoard); err != nil {
		return nil, err
	}
	origRounds, err := orig.Rounds()
	if err != nil {
		return nil, err
	}
	rounds, _ := botapi.NewReplay_Round_List(seg, int32(origRounds.Len())+1)
	for i := 0; i < origRounds.Len(); i++ {
		if err := rounds.Set(i, origRounds.At(i)); err != nil {
			return nil, err
		}
	}
	if err := rounds.Set(rounds.Len()-1, round); err != nil {
		return nil, err
	}
	newReplay.SetRounds(rounds)
	return newMsg, nil
}

func (db *dbImpl) lookupGame(id gameID) (botapi.Replay, error) {
	var r botapi.Replay
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(GameBucket)
		data := b.Get([]byte(id))
		if len(data) == 0 {
			return errGameNotFound
		}

		msg, err := capnp.Unmarshal(copyBytes(data))
		if err != nil {
			return err
		}
		r, err = botapi.ReadRootReplay(msg)
		return err
	})

	return r, err
}

func (db *dbImpl) lookupGameInfo(id gameID) (*gameInfo, error) {
	var info *gameInfo
	err := db.View(func(tx *bolt.Tx) error {
		g, err := game(tx, id)
		info = g
		return err
	})
	return info, err
}

// When we finish a game, we want to increment the win count of the winner and
// the lose count of the loser, and make a game info entry
func (db *dbImpl) finishGame(id gameID, p1, p2 *aiInfo, w WinType) error {
	err := db.Update(func(tx *bolt.Tx) error {
		p1Stat, err := aiStat(tx, p1.ID)
		if err != nil {
			return err
		}

		p2Stat, err := aiStat(tx, p2.ID)
		if err != nil {
			return err
		}

		switch w {
		case P1Win:
			p1Stat.Wins++
			p2Stat.Losses++
		case P2Win:
			p1Stat.Losses++
			p2Stat.Wins++
		case Tie:
			p1Stat.Ties++
			p2Stat.Ties++
		}

		if err := writeAiStats(tx, p1.ID, p1Stat); err != nil {
			return err
		}
		if err := writeAiStats(tx, p2.ID, p2Stat); err != nil {
			return err
		}
		g := &gameInfo{
			AI1: p1,
			AI2: p2,
		}
		return writeGameInfo(tx, id, g)
	})
	return err
}

func copyBytes(b []byte) []byte {
	bb := make([]byte, len(b))
	copy(bb, b)
	return bb
}

// All the functions below here are only meant to be called from within a transaction
func user(tx *bolt.Tx, a accessToken) (*userInfo, error) {
	var u userInfo
	b := tx.Bucket(UserBucket)
	dat := b.Get([]byte(a))
	log.Printf("LoadUser: Loading UserBucket at [%s]=%v", a, dat)
	if len(dat) == 0 {
		return nil, errUserNotFound
	}

	buf := bytes.NewReader(dat)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&u); err != nil {
		return &u, err
	}
	log.Printf("LoadUser: Decoding from UserBucket at [%s]=%v", a, u)

	return &u, nil
}

func ai(tx *bolt.Tx, id aiID) (*aiInfo, error) {
	var a aiInfo
	b := tx.Bucket(AIBucket)
	dat := b.Get([]byte(id))
	log.Printf("LookupAI: Loading AIBucket at [%s]=%v", id, dat)
	if len(dat) == 0 {
		return nil, errAINotFound
	}

	buf := bytes.NewReader(dat)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&a); err != nil {
		return &a, err
	}
	log.Printf("LookupAI: Decoding from AIBucket at [%s]=%v", id, a)

	return &a, nil
}

func aiStat(tx *bolt.Tx, id aiID) (*aiStats, error) {
	var a aiStats
	b := tx.Bucket(AIStatsBucket)
	dat := b.Get([]byte(id))
	log.Printf("LookupAI: Loading AIStatsBucket at [%s]=%v", id, dat)
	if len(dat) == 0 {
		return nil, errAINotFound
	}

	buf := bytes.NewReader(dat)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&a); err != nil {
		return &a, err
	}
	log.Printf("LookupAI: Decoding from AIStatsBucket at [%s]=%v", id, a)

	return &a, nil
}

func game(tx *bolt.Tx, id gameID) (*gameInfo, error) {
	var g gameInfo
	b := tx.Bucket(GameInfoBucket)
	dat := b.Get([]byte(id))
	log.Printf("LookupGameInfo: Loading GameInfoBucket at [%s]=%v", id, dat)
	if len(dat) == 0 {
		return nil, errGameNotFound
	}

	buf := bytes.NewReader(dat)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(&g); err != nil {
		return &g, err
	}
	log.Printf("LookupGameInfo: Decoding from GameInfoBucket at [%s]=%v", id, g)

	return &g, nil
}

func writeAi(tx *bolt.Tx, info *aiInfo) (aiID, error) {
	// Save the AI to the AIBucket
	b := tx.Bucket(AIBucket)
	var id aiID
	if info.ID == aiID("") {
		idNum, err := b.NextSequence()
		if err != nil {
			return id, err
		}
		id = aiID(strconv.FormatUint(idNum, 10))
	} else {
		id = info.ID
	}
	info.ID = id
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(info); err != nil {
		return id, err
	}
	log.Printf("CreateAI: Putting into AIBucket at [%s]=%v", id, buf.Bytes())
	return id, b.Put([]byte(id), buf.Bytes())
}

func writeUser(tx *bolt.Tx, info *userInfo) (uID, error) {
	// Save the user to the UserBucket
	b := tx.Bucket(UserBucket)
	var id uID
	if info.ID == uID("") {
		idNum, err := b.NextSequence()
		if err != nil {
			return id, err
		}
		id = uID(strconv.FormatUint(idNum, 10))
	} else {
		id = info.ID
	}
	info.ID = id
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(info); err != nil {
		return id, err
	}
	log.Printf("CreateUser: Putting into UserBucket at [%s]=%v", info.Token, buf.Bytes())
	return id, b.Put([]byte(info.Token), buf.Bytes())
}

func writeAiStats(tx *bolt.Tx, id aiID, stats *aiStats) error {
	// Save the AIStats to the AIStatsBucket
	b := tx.Bucket(AIStatsBucket)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(stats); err != nil {
		return err
	}
	log.Printf("CreateAIStats: Putting into AIStatsBucket at [%s]=%v", id, buf.Bytes())
	return b.Put([]byte(id), buf.Bytes())
}

func writeGameInfo(tx *bolt.Tx, id gameID, info *gameInfo) error {
	// Save the GameInfo to the GameInfoBucket
	b := tx.Bucket(GameInfoBucket)
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(info); err != nil {
		return err
	}
	log.Printf("CreateGameInfo: Putting into GameInfoBucket at [%s]=%v", id, buf.Bytes())
	return b.Put([]byte(id), buf.Bytes())
}
