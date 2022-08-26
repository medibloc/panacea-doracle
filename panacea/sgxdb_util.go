package panacea

import (
	"bytes"
	"encoding/gob"
	"github.com/medibloc/panacea-doracle/store/sgxleveldb"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func SaveTrustedBlockInfo(info TrustedBlockInfo) error {
	var buffer bytes.Buffer

	enc := gob.NewEncoder(&buffer)

	err := enc.Encode(info)
	if err != nil {
		return err
	}

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dbDir := filepath.Join(userHomeDir, ".doracle", "data")

	db, err := sgxleveldb.NewSgxLevelDB("light-client-db", dbDir)
	if err != nil {
		return err
	}
	err = db.Set([]byte("trustedBlockInfo"), buffer.Bytes())
	if err != nil {
		return err
	}

	err = db.Close()
	if err != nil {
		return err
	}

	log.Info("Save trusted block info to levelDB")
	return nil

}

func GetTrustedBlockInfo() (TrustedBlockInfo, error) {

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return TrustedBlockInfo{}, err
	}
	dbDir := filepath.Join(userHomeDir, ".doracle", "data")

	db, err := sgxleveldb.NewSgxLevelDB("light-client-db", dbDir)
	if err != nil {
		return TrustedBlockInfo{}, err
	}

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

	log.Info("Save trusted block info to levelDB")

	return info, nil
}
