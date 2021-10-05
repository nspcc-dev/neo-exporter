package locode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/nspcc-dev/neofs-net-monitor/pkg/morphchain"
	"go.etcd.io/bbolt"
)

var errNotExist = errors.New("database file does not exists")

// Open opens underlying BoltDB instance.
//
// Timeout of BoltDB opening is 3s (only for Linux or Darwin).
func (db *DB) Open() error {
	_, err := os.Stat(db.path)
	if os.IsNotExist(err) {
		return errNotExist
	}

	db.bolt, err = bbolt.Open(db.path, db.mode, db.boltOpts)
	if err != nil {
		return fmt.Errorf("could not open BoltDB: %w", err)
	}

	return nil
}

// Close closes underlying BoltDB instance.
//
// Must not be called before successful Open call.
func (db *DB) Close() error {
	if db.bolt != nil {
		return db.bolt.Close()
	}

	return nil
}

type recordJSON struct {
	LocationName string  `json:"LocationName"`
	Lat          float64 `json:"Latitude"`
	Long         float64 `json:"Longitude"`
}

func (r recordJSON) Latitude() float64 {
	return r.Lat
}

func (r recordJSON) Longitude() float64 {
	return r.Long
}

func (r recordJSON) Location() string {
	return r.LocationName
}

var (
	errRecordNotFound   = errors.New("could not get record by provided locode")
	errDBNotInitialized = errors.New("DB instance not initialized")
)

type (
	countryCode  = [2]byte
	locationCode = [3]byte

	Position interface {
		Latitude() float64
		Longitude() float64
		Location() string
	}
)

func (db *DB) Get(node *morphchain.Node) (position Position, err error) {
	if db.bolt == nil {
		return nil, errDBNotInitialized
	}

	cc, lc, err := parseLocode(node.Locode)
	if err != nil {
		return nil, err
	}

	err = db.bolt.View(func(tx *bbolt.Tx) error {
		bktCountry := tx.Bucket(cc[:])
		if bktCountry == nil {
			return errRecordNotFound
		}

		data := bktCountry.Get(lc[:])
		if data == nil {
			return errRecordNotFound
		}

		rj := new(recordJSON)

		err := json.Unmarshal(data, rj)
		if err != nil {
			return err
		}

		position = rj

		return err
	})

	return
}

var errIncorrectLocode = errors.New("incorrect locode")

func parseLocode(l string) (cc countryCode, lc locationCode, err error) {
	locParts := strings.Split(l, " ")

	if len(locParts) != 2 {
		return cc, lc, errIncorrectLocode
	}

	if len(locParts[0]) != 2 {
		return cc, lc, errIncorrectLocode
	}

	copy(cc[:], locParts[0])

	if len(locParts[1]) != 3 {
		return cc, lc, errIncorrectLocode
	}

	copy(lc[:], locParts[1])

	return
}
