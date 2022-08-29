package panacea

import (
	"bytes"
	"encoding/gob"
	"github.com/medibloc/panacea-doracle/store/sgxleveldb"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var DbDir string

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DbDir = filepath.Join(userHomeDir, ".doracle", "data")
}

func SaveTrustedBlockInfo(info TrustedBlockInfo) error {
	var buffer bytes.Buffer

	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(info)
	if err != nil {
		return err
	}

	db, err := sgxleveldb.NewSgxLevelDB("light-client-db", DbDir)
	if err != nil {
		return err
	}

	defer db.Close()

	err = db.Set([]byte("trustedBlockInfo"), buffer.Bytes())
	if err != nil {
		return err
	}

	log.Info("Save trusted block info to levelDB")
	return nil

}

func GetTrustedBlockInfo() (TrustedBlockInfo, error) {

	db, err := sgxleveldb.NewSgxLevelDB("light-client-db", DbDir)
	if err != nil {
		return TrustedBlockInfo{}, err
	}

	defer db.Close()

	infoBuf, err := db.Get([]byte("trustedBlockInfo"))
	if err != nil {
		return TrustedBlockInfo{}, err
	}

	var buffer = bytes.NewBuffer(infoBuf)
	dec := gob.NewDecoder(buffer)

	var info TrustedBlockInfo

	err = dec.Decode(&info)
	if err != nil {
		return TrustedBlockInfo{}, err
	}

	log.Info("Get trusted block info to levelDB")

	return info, nil
}
