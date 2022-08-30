package sgxleveldb

import (
	"github.com/medibloc/panacea-doracle/sgx"
	log "github.com/sirupsen/logrus"
	tmdb "github.com/tendermint/tm-db"
)

type sgxLevelDBBatch struct {
	tmdb.Batch
}

func (sbatch *sgxLevelDBBatch) Set(key, value []byte) error {
	log.Info("ENCRYPT SOMETHING SECRETLY - BATCH")
	sealValue, err := sgx.Seal(value, true)
	if err != nil {
		return err
	}
	return sbatch.Batch.Set(key, sealValue)
}
