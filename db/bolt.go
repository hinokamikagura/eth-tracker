package db

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/bits-and-blooms/bitset"
	bolt "go.etcd.io/bbolt"
)

type boltDB struct {
	db *bolt.DB
}

const (
	dbFolderName = "db/tracker_db"

	upperBoundKey = "upper"
	lowerBoundKey = "lower"
)

var sortableOrder = binary.BigEndian

func NewBoltDB() (DB, error) {
	db, err := bolt.Open(dbFolderName, 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("blocks"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	return &boltDB{
		db: db,
	}, nil
}

func (d *boltDB) Close() error {
	return d.db.Close()
}

func (d *boltDB) get(k string) ([]byte, error) {
	var v []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		v = b.Get([]byte(k))
		return nil
	})

	if err != nil {
		return nil, err
	}

	return v, nil
}

func (d *boltDB) setUint64(k string, v uint64) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		return b.Put([]byte(k), marshalUint64(v))
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *boltDB) setUint64AsKey(v uint64) error {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		return b.Put(marshalUint64(v), nil)
	})
	if err != nil {
		return err
	}
	return nil
}

func unmarshalUint64(b []byte) uint64 {
	return sortableOrder.Uint64(b)
}

func marshalUint64(v uint64) []byte {
	b := make([]byte, 8)
	sortableOrder.PutUint64(b, v)
	return b
}

func (d *boltDB) SetLowerBound(v uint64) error {
	return d.setUint64(lowerBoundKey, v)
}

func (d *boltDB) GetLowerBound() (uint64, error) {
	v, err := d.get(lowerBoundKey)
	if err != nil {
		return 0, err
	}

	if v == nil {
		return 0, nil
	}

	return unmarshalUint64(v), nil
}

func (d *boltDB) SetUpperBound(v uint64) error {
	return d.setUint64(upperBoundKey, v)
}

func (d *boltDB) GetUpperBound() (uint64, error) {
	v, err := d.get(upperBoundKey)
	if err != nil {
		return 0, err
	}
	return unmarshalUint64(v), nil
}

func (d *boltDB) SetValue(v uint64) error {
	return d.setUint64AsKey(v)
}

func (d *boltDB) GetMissingValuesBitSet(lowerBound uint64, upperBound uint64) (*bitset.BitSet, error) {
	var b bitset.BitSet

	err := d.db.View(func(tx *bolt.Tx) error {
		var (
			lowerRaw = marshalUint64(lowerBound)
			upperRaw = marshalUint64(upperBound)
		)

		for i := lowerBound; i <= upperBound; i++ {
			b.Set(uint(i))
		}

		c := tx.Bucket([]byte("blocks")).Cursor()

		for k, _ := c.Seek(lowerRaw); k != nil && bytes.Compare(k, upperRaw) <= 0; k, _ = c.Next() {
			b.Clear(uint(unmarshalUint64(k)))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (d *boltDB) Cleanup() error {
	lowerBound, err := d.GetLowerBound()
	if err != nil {
		return err
	}
	target := marshalUint64(lowerBound - 1)

	err = d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("blocks"))
		c := b.Cursor()

		for k, _ := c.First(); k != nil && bytes.Compare(k, target) <= 0; k, _ = c.Next() {
			if err := b.Delete(k); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
