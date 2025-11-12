package db

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/bits-and-blooms/bitset"
	bolt "go.etcd.io/bbolt"
)

const (
	// dbFolderName is the path where the BoltDB file is stored
	dbFolderName = "db/tracker_db"
	// dbFileMode is the file mode for the database file (read-write for owner only)
	dbFileMode = 0600
	// bucketName is the name of the BoltDB bucket for storing block data
	bucketName = "blocks"
	// upperBoundKey is the key for storing the upper bound block number
	upperBoundKey = "upper"
	// lowerBoundKey is the key for storing the lower bound block number
	lowerBoundKey = "lower"
)

var (
	// sortableOrder is the byte order used for encoding/decoding uint64 values
	// Using BigEndian ensures lexicographic ordering matches numeric ordering
	sortableOrder = binary.BigEndian
)

// boltDB implements the DB interface using BoltDB for persistent storage.
type boltDB struct {
	db *bolt.DB
}

// NewBoltDB creates a new BoltDB instance and initializes the required bucket.
func NewBoltDB() (DB, error) {
	db, err := bolt.Open(dbFolderName, dbFileMode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize the blocks bucket
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
		return nil
	}); err != nil {
		db.Close()
		return nil, err
	}

	return &boltDB{
		db: db,
	}, nil
}

func (d *boltDB) Close() error {
	return d.db.Close()
}

// get retrieves a value from the database by key.
func (d *boltDB) get(k string) ([]byte, error) {
	var v []byte
	err := d.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		v = b.Get([]byte(k))
		return nil
	})

	return v, err
}

// setUint64 stores a uint64 value with the given key.
func (d *boltDB) setUint64(k string, v uint64) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.Put([]byte(k), marshalUint64(v))
	})
}

// setUint64AsKey stores a block number as a key (for efficient range queries).
func (d *boltDB) setUint64AsKey(v uint64) error {
	return d.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}
		return b.Put(marshalUint64(v), nil)
	})
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

// GetMissingValuesBitSet returns a bitset indicating which blocks in the range are missing.
// The bitset is initialized with all blocks set, then cleared for blocks that exist in the database.
func (d *boltDB) GetMissingValuesBitSet(lowerBound uint64, upperBound uint64) (*bitset.BitSet, error) {
	var b bitset.BitSet

	err := d.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		lowerRaw := marshalUint64(lowerBound)
		upperRaw := marshalUint64(upperBound)

		// Initialize bitset: set all blocks in range as missing
		for i := lowerBound; i <= upperBound; i++ {
			b.Set(uint(i))
		}

		// Clear bits for blocks that exist in the database
		c := bucket.Cursor()
		for k, _ := c.Seek(lowerRaw); k != nil && bytes.Compare(k, upperRaw) <= 0; k, _ = c.Next() {
			// Only process block number keys (8 bytes), not metadata keys
			if len(k) == 8 {
				b.Clear(uint(unmarshalUint64(k)))
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &b, nil
}

// Cleanup removes block entries that are below the current lower bound.
// This helps prevent unbounded database growth.
func (d *boltDB) Cleanup() error {
	lowerBound, err := d.GetLowerBound()
	if err != nil {
		return err
	}

	if lowerBound == 0 {
		return nil // Nothing to clean up
	}

	target := marshalUint64(lowerBound - 1)

	return d.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return fmt.Errorf("bucket %s not found", bucketName)
		}

		c := bucket.Cursor()
		for k, _ := c.First(); k != nil && bytes.Compare(k, target) <= 0; k, _ = c.Next() {
			// Only delete block number keys (8 bytes), not metadata keys
			if len(k) == 8 {
				if err := bucket.Delete(k); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
