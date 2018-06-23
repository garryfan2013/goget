package engine

import (
	"time"

	"github.com/garryfan2013/goget/store"

	"github.com/boltdb/bolt"
)

const (
	dbName      = "goget.db"
	openTimeOut = 3
)

func init() {
	store.Register(&BoltDBStoreBuilder{})
}

type BoltDBStoreBuilder struct {
}

func (b *BoltDBStoreBuilder) Build() (store.Storer, error) {
	db, err := bolt.Open(dbName, 0600, &bolt.Options{Timeout: openTimeOut * time.Second})
	if err != nil {
		return nil, err
	}

	return &BoltDBStore{db}, nil
}

func (b *BoltDBStoreBuilder) Name() string {
	return store.StoreBoltDB
}

type BoltDBStore struct {
	db *bolt.DB
}

func (s *BoltDBStore) CreateBucket(bucket string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *BoltDBStore) Put(bucket string, key string, value []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}
		return b.Put([]byte(key), value)
	})
}

func (s *BoltDBStore) Get(bucket string, key string) ([]byte, error) {
	var val []byte
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		val = b.Get([]byte(key))
		return nil
	}); err != nil {
		return nil, err
	}

	return val, nil
}

func (s *BoltDBStore) GetSwap(bucket string, key string, fn func([]byte) ([]byte, error)) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		val := b.Get([]byte(key))
		val, err := fn(val)
		if err != nil {
			return err
		}

		return b.Put([]byte(key), val)
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *BoltDBStore) Delete(bucket string, key string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.Delete([]byte(key))
	})
}

func (s *BoltDBStore) ForEach(bucket string, fn func(k string, v []byte) error) error {
	return s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		return b.ForEach(func(key, val []byte) error {
			return fn(string(key), val)
		})
	})
}

func (s *BoltDBStore) Open() error {
	return nil
}

func (s *BoltDBStore) Close() error {
	return s.db.Close()
}
