package db

import (
	"log/slog"

	"github.com/bits-and-blooms/bitset"
)

type (
	DB interface {
		Close() error
		GetLowerBound() (uint64, error)
		SetLowerBound(v uint64) error
		SetUpperBound(uint64) error
		GetUpperBound() (uint64, error)
		SetValue(uint64) error
		GetMissingValuesBitSet(uint64, uint64) (*bitset.BitSet, error)
		Cleanup() error
	}

	DBOpts struct {
		Logg   *slog.Logger
		DBType string
	}
)

func New(o DBOpts) (DB, error) {
	var (
		err error
		db  DB
	)

	switch o.DBType {
	case "bolt":
		db, err = NewBoltDB()
		if err != nil {
			return nil, err
		}
	default:
		db, err = NewBoltDB()
		if err != nil {
			return nil, err
		}
		o.Logg.Warn("invalid db type, using default type (bolt)")
	}

	return db, nil
}
