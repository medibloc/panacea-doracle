// Package leveldb implements a light client level db for panacea-doracle.
// It does include Set & Get functions that are sealed & unsealed in the sgx environment.

package sgxleveldb

import (
	"fmt"
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb/opt"
	tmdb "github.com/tendermint/tm-db"
)

type SgxLevelDB struct {
	*tmdb.GoLevelDB
}

func NewSgxLevelDB(name string, dir string) (*SgxLevelDB, error) {
	return NewSgxLevelDBWithOpts(name, dir, nil)
}

func NewSgxLevelDBWithOpts(name string, dir string, o *opt.Options) (*SgxLevelDB, error) {
	goLevelDB, err := tmdb.NewGoLevelDBWithOpts(name, dir, o)
	if err != nil {
		return nil, fmt.Errorf("failed to NewGoLevelDBWithOpts: %w", err)
	}

	return &SgxLevelDB{goLevelDB}, nil
}

func (sdb *SgxLevelDB) Set(key, value []byte) error {
	log.Info("ENCRYPT SOMETHING SECRETLY")
	sealValue, err := sgx.Seal(value, true)
	if err != nil {
		return err
	}
	return sdb.GoLevelDB.Set(key, sealValue)
}

func (sdb *SgxLevelDB) Get(key []byte) ([]byte, error) {
	val, err := sdb.GoLevelDB.Get(key)
	if err != nil {
		return nil, err
	}
	log.Info("DECRYPT SOMETHING SECRETLY")
	unsealedVal, err := sgx.Unseal(val, true)
	if err != nil {
		return nil, err
	}

	return unsealedVal, nil
}

func (sdb *SgxLevelDB) NewBatch() tmdb.Batch {
	batch := sdb.GoLevelDB.NewBatch()
	return &sgxLevelDBBatch{batch}
}