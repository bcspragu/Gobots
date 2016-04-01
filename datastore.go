package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"strconv"
	"time"

	"zombiezen.com/go/capnproto2"

	"github.com/bcspragu/Gobots/botapi"
	"github.com/boltdb/bolt"
)

type datastore interface {
	// Users
	createUser(u userInfo) (id uID, err error)
	loadUser(a accessToken) (*userInfo, error)
	userExists(name string) (bool, error)

	// AIs
	createAI(info *aiInfo, a accessToken) (id aiID, err error)
	listAIsForUser(a accessToken) ([]*aiInfo, error)
	lookupAI(id aiID) (*aiInfo, error)

	// Games
	startGame(ai1, ai2 aiID, init botapi.Board) (gameID, error)
	addRound(id gameID, round botapi.Replay_Round) error
	lookupGame(id gameID) (botapi.Replay, error)
}

type dbImpl struct {
	*bolt.DB
}

func initDB(dbName string) (datastore, error) {
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range [][]byte{UserBucket, UserLookupBucket, GameBucket, AIBucket} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}

		return nil
	})

	return &dbImpl{db}, err
}

type accessToken string

type uID string

type aiID string

type gameID string

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

	Wins   int
	Losses int
}

var (
	AIBucket = []byte("AI") // aID -> aiInfo

	GameBucket = []byte("Games") //

	UserBucket       = []byte("Users")       // accessToken -> userInfo
	UserLookupBucket = []byte("UserLookups") // userInfo.Name -> []byte{}
)

func (db *dbImpl) createUser(uInfo userInfo) (id uID, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		idNum, err := b.NextSequence()
		if err != nil {
			return err
		}

		id = uID(strconv.FormatUint(idNum, 10))
		uInfo.ID = id
		log.Printf("CreateUser: Creating user named %s with ID %s", uInfo.Name, uInfo.ID)
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(uInfo); err != nil {
			return err
		}
		log.Printf("CreateUser: Putting into UserBucket [%s]=%v", uInfo.Token, buf.Bytes())
		if err := b.Put([]byte(uInfo.Token), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserBucket: " + err.Error())
		}
		buf = bytes.Buffer{}

		b = tx.Bucket(UserLookupBucket)
		log.Printf("CreateUser: Putting empty bytes UserLookupBucket [%s]=[]byte{}", uInfo.Name)
		if err := b.Put([]byte(uInfo.Name), []byte{}); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserLookupBucket: " + err.Error())
		}
		return nil
	})
	return
}

func (db *dbImpl) loadUser(a accessToken) (*userInfo, error) {
	var u userInfo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		dat := b.Get([]byte(a))
		log.Printf("LoadUser: Loading UserBucket at [%s]=%v", a, dat)
		if len(dat) == 0 {
			return errUserNotFound
		}

		buf := bytes.NewReader(dat)
		dec := gob.NewDecoder(buf)

		return dec.Decode(&u)
	})
	log.Printf("LoadUser: Decoding from UserBucket at [%s]=%v", a, u)

	return &u, err
}

func (db *dbImpl) userExists(name string) (exists bool, err error) {
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
		// Load the user first to get their ID
		var u *userInfo
		b := tx.Bucket(UserBucket)
		dat := b.Get([]byte(a))
		log.Printf("CreateAI: Loading UserBucket at [%s]=%v", a, dat)
		if len(dat) == 0 {
			return errUserNotFound
		}
		buf := bytes.NewBuffer(dat)
		if err := gob.NewDecoder(buf).Decode(&u); err != nil {
			return err
		}
		log.Printf("CreateAI: Decoding from UserBucket at [%s]=%v", a, u)
		buf = new(bytes.Buffer)

		for _, ai := range u.AIs {
			if ai.Name == info.Name {
				log.Printf("CreateAI: Bot %s already exists, not creating", info.Name)
				return errors.New("Bot already exists")
			}
		}

		// Save the AI to the AIBucket
		b = tx.Bucket(AIBucket)
		idNum, err := b.NextSequence()
		if err != nil {
			return err
		}
		id = aiID(strconv.FormatUint(idNum, 10))
		newInfo := &aiInfo{
			ID:     id,
			Name:   info.Name,
			UserID: u.ID,
		}
		if err := gob.NewEncoder(buf).Encode(newInfo); err != nil {
			return err
		}
		log.Printf("CreateAI: Putting into AIBucket at [%s]=%v", id, buf.Bytes())
		if err := b.Put([]byte(id), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to AIBucket: " + err.Error())
		}
		log.Println("put", buf.Bytes(), "in id ", id)
		buf = new(bytes.Buffer)

		b = tx.Bucket(UserBucket)
		// Append the AI to the list of AIs for this user and save that
		u.AIs = append(u.AIs, newInfo)
		if err := gob.NewEncoder(buf).Encode(u); err != nil {
			return err
		}
		log.Printf("CreateAI: Putting into UserBucket at [%s]=%v", a, buf.Bytes())
		if err := b.Put([]byte(a), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserAIBucket: " + err.Error())
		}
		return nil
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
		b := tx.Bucket(AIBucket)
		dat := b.Get([]byte(id))
		if len(dat) == 0 {
			return errAINotFound
		}

		buf := bytes.NewReader(dat)
		return gob.NewDecoder(buf).Decode(&info)
	})
	return info, err
}

// Games
func (db *dbImpl) startGame(ai1, ai2 aiID, init botapi.Board) (gameID, error) {
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
		r.SetInitialBoard(init)

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
	initBoard, err := orig.InitialBoard()
	if err != nil {
		return nil, err
	}
	if err := newReplay.SetInitialBoard(initBoard); err != nil {
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

func copyBytes(b []byte) []byte {
	bb := make([]byte, len(b))
	copy(bb, b)
	return bb
}

var errDatastoreNotImplemented = errors.New("gobots: datastore operation not implemented")
var errUserNotFound = errors.New("gobots: user not found")
var errAINotFound = errors.New("gobots: AI not found")
var errGameNotFound = errors.New("gobots: game not found")
