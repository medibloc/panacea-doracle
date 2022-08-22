// Package sgxLevelDB implements a light client level db for panacea-doracle.
// It does include Set & Get functions that are sealed & unsealed in the sgx environment.

package sgxLevelDB

import (
	"fmt"
	"path/filepath"

	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	tmdb "github.com/tendermint/tm-db"
)

var (
	// errBatchClosed is returned when a closed or written batch is used.
	errBatchClosed = errors.New("batch has been written or closed")

	// errKeyEmpty is returned when attempting to use an empty or nil key.
	errKeyEmpty = errors.New("key cannot be empty")

	// errValueNil is returned when attempting to set a nil value.
	errValueNil = errors.New("value cannot be nil")
)

type dbCreator func(name string, dir string) (tmdb.DB, error)

var backends = map[tmdb.BackendType]dbCreator{}

func registerDBCreator(backend tmdb.BackendType, creator dbCreator, force bool) {
	_, ok := backends[backend]
	if !force && ok {
		return
	}
	backends[backend] = creator
}

func init() {
	dbCreator := func(name string, dir string) (tmdb.DB, error) {
		return NewGoLevelDB(name, dir)
	}
	registerDBCreator(tmdb.GoLevelDBBackend, dbCreator, false)
}

type GoLevelDB struct {
	Db *leveldb.DB
}

var _ tmdb.DB = (*GoLevelDB)(nil)

func NewGoLevelDB(name string, dir string) (*GoLevelDB, error) {
	return NewGoLevelDBWithOpts(name, dir, nil)
}

func NewGoLevelDBWithOpts(name string, dir string, o *opt.Options) (*GoLevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, o)
	if err != nil {
		return nil, err
	}
	database := &GoLevelDB{
		Db: db,
	}
	return database, nil
}

// Get implements DB.
func (db *GoLevelDB) Get(key []byte) ([]byte, error) {
	if len(key) == 0 {
		return nil, errKeyEmpty
	}
	res, err := db.Db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	// unseal data
	unsealedRes, err := sgx.Unseal(res, true)
	log.Debug("Unseal value from key : %x\n", key)
	if err != nil {
		return nil, err
	}

	return unsealedRes, nil
}

// Has implements DB.
func (db *GoLevelDB) Has(key []byte) (bool, error) {
	bytes, err := db.Get(key)
	if err != nil {
		return false, err
	}
	return bytes != nil, nil
}

// Set implements DB.
func (db *GoLevelDB) Set(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	// seal value
	sealValue, err := sgx.Seal(value, true)
	log.Debug("Seal value of key: %x\n", key)

	if err != nil {
		return err
	}
	if err := db.Db.Put(key, sealValue, nil); err != nil {
		return err
	}
	return nil
}

// SetSync implements DB.
func (db *GoLevelDB) SetSync(key []byte, value []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	// seal value
	sealValue, err := sgx.Seal(value, true)
	log.Debug("Seal value of key: %x\n", key)

	if err != nil {
		return err
	}
	if err := db.Db.Put(key, sealValue, &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	return nil
}

// Delete implements DB.
func (db *GoLevelDB) Delete(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	if err := db.Db.Delete(key, nil); err != nil {
		return err
	}
	return nil
}

// DeleteSync implements DB.
func (db *GoLevelDB) DeleteSync(key []byte) error {
	if len(key) == 0 {
		return errKeyEmpty
	}
	err := db.Db.Delete(key, &opt.WriteOptions{Sync: true})
	if err != nil {
		return err
	}
	return nil
}

func (db *GoLevelDB) DB() *leveldb.DB {
	return db.Db
}

// Close implements DB.
func (db *GoLevelDB) Close() error {
	if err := db.Db.Close(); err != nil {
		return err
	}
	return nil
}

// Print implements DB.
func (db *GoLevelDB) Print() error {
	str, err := db.Db.GetProperty("leveldb.stats")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", str)

	itr := db.Db.NewIterator(nil, nil)
	for itr.Next() {
		key := itr.Key()
		value := itr.Value()
		// unseal value
		unsealValue, err := sgx.Unseal(value, true)
		log.Debug("Unseal value from key : %x\n", key)

		if err != nil {
			return err
		}
		fmt.Printf("[%X]:\t[%X]\n", key, unsealValue)
	}
	return nil
}

// Stats implements DB.
func (db *GoLevelDB) Stats() map[string]string {
	keys := []string{
		"leveldb.num-files-at-level{n}",
		"leveldb.stats",
		"leveldb.sstables",
		"leveldb.blockpool",
		"leveldb.cachedblock",
		"leveldb.openedtables",
		"leveldb.alivesnaps",
		"leveldb.aliveiters",
	}

	stats := make(map[string]string)
	for _, key := range keys {
		str, err := db.Db.GetProperty(key)
		if err == nil {
			stats[key] = str
		}
	}
	return stats
}

func (db *GoLevelDB) ForceCompact(start, limit []byte) error {
	return db.Db.CompactRange(util.Range{Start: start, Limit: limit})
}

// NewBatch implements DB.
func (db *GoLevelDB) NewBatch() tmdb.Batch {
	return newGoLevelDBBatch(db)
}

// Iterator implements DB.
func (db *GoLevelDB) Iterator(start, end []byte) (tmdb.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	itr := db.Db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	return newGoLevelDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements DB.
func (db *GoLevelDB) ReverseIterator(start, end []byte) (tmdb.Iterator, error) {
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	itr := db.Db.NewIterator(&util.Range{Start: start, Limit: end}, nil)
	return newGoLevelDBIterator(itr, start, end, true), nil
}
