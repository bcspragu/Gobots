package main

import (
	"bytes"
	"encoding/gob"
	"errors"
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
		for _, b := range [][]byte{UserBucket, UserLookupBucket, GameBucket, UserAIBucket, AIBucket} {
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
}

type aiInfo struct {
	ID     aiID
	Name   string
	UserID uID

	Wins   int
	Losses int
}

var (
	AIBucket     = []byte("AI")     // aID -> aiInfo
	UserAIBucket = []byte("UserAI") // accessToken -> []aiInfo

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
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(uInfo); err != nil {
			return err
		}
		if err := b.Put([]byte(uInfo.Token), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserBucket: " + err.Error())
		}
		buf.Reset()

		b = tx.Bucket(UserLookupBucket)
		if err := b.Put([]byte(uInfo.Name), []byte{}); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserLookupBucket: " + err.Error())
		}

		b = tx.Bucket(UserAIBucket)
		if err := gob.NewEncoder(&buf).Encode([]*aiInfo{}); err != nil {
			return err
		}
		if err := b.Put([]byte(uInfo.Token), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserAIBucket: " + err.Error())
		}
		return nil
	})
	return
}

func (db *dbImpl) loadUser(a accessToken) (*userInfo, error) {
	var u *userInfo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserBucket)
		dat := b.Get([]byte(a))

		buf := bytes.NewReader(dat)
		dec := gob.NewDecoder(buf)

		return dec.Decode(&u)
	})

	return u, err
}

func (db *dbImpl) userExists(name string) (exists bool, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserLookupBucket)
		// If the bucket has that key, the user exists
		exists = b.Get([]byte(name)) != nil
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
		if len(dat) == 0 {
			return errDatastoreNotFound
		}
		buf := bytes.NewBuffer(dat)
		if err := gob.NewDecoder(buf).Decode(&u); err != nil {
			return err
		}
		buf.Reset()

		// Load the user's existing AIs
		var ais []*aiInfo
		b = tx.Bucket(UserAIBucket)
		dat = b.Get([]byte(a))
		if len(dat) == 0 {
			ais = []*aiInfo{}
		} else {
			if _, err := buf.Write(dat); err != nil {
				return err
			}
			if err := gob.NewDecoder(buf).Decode(&ais); err != nil {
				return err
			}
			buf.Reset()
		}

		for _, ai := range ais {
			if ai.Name == info.Name {
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
		if err := b.Put([]byte(id), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to AIBucket: " + err.Error())
		}
		buf.Reset()

		// Append the AI to the list of AIs for this user and save that
		b = tx.Bucket(UserAIBucket)
		ais = append(ais, newInfo)
		if err := gob.NewEncoder(buf).Encode(ais); err != nil {
			return err
		}
		if err := b.Put([]byte(u.ID), buf.Bytes()); err != nil {
			return errors.New("DB might be in an inconsistent state, failed writing to UserAIBucket: " + err.Error())
		}
		return nil
	})
	return
}

func (db *dbImpl) listAIsForUser(a accessToken) ([]*aiInfo, error) {
	var infos []*aiInfo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(UserAIBucket)
		dat := b.Get([]byte(a))
		if len(dat) == 0 {
			infos = []*aiInfo{}
			return nil
		}

		buf := bytes.NewReader(dat)
		return gob.NewDecoder(buf).Decode(&infos)
	})
	return infos, err
}

func (db *dbImpl) lookupAI(id aiID) (*aiInfo, error) {
	var info *aiInfo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(AIBucket)
		dat := b.Get([]byte(id))
		if len(dat) == 0 {
			return errDatastoreNotFound
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
			return errDatastoreNotFound
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
			return errDatastoreNotFound
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
var errDatastoreNotFound = errors.New("gobots: datastore entity not found")
