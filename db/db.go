// Package db provides database interfaces and implementations for block state persistence.
// It tracks processed blocks and identifies gaps for backfilling.
package db

import (
	"log/slog"

	"github.com/bits-and-blooms/bitset"
)

const (
	// DBTypeBolt is the BoltDB implementation type
	DBTypeBolt = "bolt"
)

// DB defines the interface for block state persistence.
type DB interface {
	// Close closes the database connection and releases resources.
	Close() error

	// GetLowerBound returns the lowest block number being tracked.
	GetLowerBound() (uint64, error)

	// SetLowerBound sets the lowest block number to track.
	SetLowerBound(v uint64) error

	// SetUpperBound sets the highest block number being tracked.
	SetUpperBound(uint64) error

	// GetUpperBound returns the highest block number being tracked.
	GetUpperBound() (uint64, error)

	// SetValue marks a block number as processed.
	SetValue(uint64) error

	// GetMissingValuesBitSet returns a bitset indicating which blocks in the range are missing.
	// Bits are set for missing blocks and cleared for processed blocks.
	GetMissingValuesBitSet(uint64, uint64) (*bitset.BitSet, error)

	// Cleanup removes old block entries below the lower bound.
	Cleanup() error
}

// DBOpts contains configuration options for creating a new DB instance.
type DBOpts struct {
	Logg   *slog.Logger // Structured logger
	DBType string       // Database implementation type ("bolt")
}

// New creates a new DB instance based on the provided options.
func New(o DBOpts) (DB, error) {
	var (
		err error
		db  DB
	)

	switch o.DBType {
	case DBTypeBolt:
		db, err = NewBoltDB()
		if err != nil {
			return nil, err
		}
	default:
		o.Logg.Warn("unknown db type, using default bolt implementation", "db_type", o.DBType)
		db, err = NewBoltDB()
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}
