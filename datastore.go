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
	createUser(u userInfo) error
	loadUser(a accessToken) (*userInfo, error)

	// AIs
	createAI(info *aiInfo) (id aiID, err error)
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
		for _, b := range [][]byte{UserBucket, GameBucket, AIBucket} {
			if _, err := tx.CreateBucketIfNotExists(b); err != nil {
				return err
			}
		}

		return nil
	})

	return &dbImpl{db}, err
}

type accessToken string

type aiID string

type gameID string

type userInfo struct {
	Name  string
	Token accessToken
}

type aiInfo struct {
	ID    aiID
	Name  string
	Token accessToken // Owner's access token

	Wins   int
	Losses int
}

var (
	UserBucket = []byte("Users")
	GameBucket = []byte("Games")
	AIBucket   = []byte("AI")
)

func (db *dbImpl) createUser(uInfo userInfo) error {
	return db.Update(func(tx *bolt.Tx) error {
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		if err := enc.Encode(uInfo); err != nil {
			return err
		}

		b := tx.Bucket(UserBucket)
		return b.Put([]byte(uInfo.Token), buf.Bytes())
	})
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

// AIs
func (db *dbImpl) createAI(info *aiInfo) (id aiID, err error) {
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(AIBucket)
		idNum, err := b.NextSequence()
		if err != nil {
			return err
		}
		id = aiID(strconv.FormatUint(idNum, 10))
		newInfo := &aiInfo{
			ID:    id,
			Name:  info.Name,
			Token: info.Token,
		}
		var buf bytes.Buffer
		if err := gob.NewEncoder(&buf).Encode(newInfo); err != nil {
			return err
		}
		return b.Put([]byte(id), buf.Bytes())
	})
	return
}

// TODO: Instead of iterating through all of the AIs looking for ones owned by
// this person, find something better
func (db *dbImpl) listAIsForUser(a accessToken) ([]*aiInfo, error) {
	var result []*aiInfo
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(AIBucket)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if len(v) == 0 {
				continue
			}
			info := new(aiInfo)
			if err := gob.NewDecoder(bytes.NewReader(v)).Decode(info); err != nil {
				continue
			}
			if info.Token == a {
				result = append(result, info)
			}
		}
		return nil
	})
	return result, err
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
