package store

import (
	"bytes"
	"encoding/json"

	shell "github.com/ipfs/go-ipfs-api"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	log "github.com/sirupsen/logrus"
)

type Ipfs struct {
	sh *shell.Shell
}

// NewIpfs generates an ipfs node with ipfs url.
func NewIpfs(url string) *Ipfs {
	newShell := shell.NewShell(url)

	log.Info("successfully connect to IPFS node")

	return &Ipfs{
		sh: newShell,
	}
}

// Add method adds a data and returns a CID.
func (i *Ipfs) Add(data []byte) (string, error) {
	reader := bytes.NewReader(data)

	cid, err := i.sh.Add(reader)
	if err != nil {
		return "", err
	}

	return cid, nil
}

// Get method gets a data and returns a data schema of Deal.
func (i *Ipfs) Get(cid string) ([]string, error) {
	data, err := i.sh.Cat(cid)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(data)
	if err != nil {
		return nil, err
	}

	newStr := buf.String()

	var deal types.Deal

	err = json.Unmarshal([]byte(newStr), &deal)
	if err != nil {
		return nil, err
	}

	return deal.DataSchema, nil
}
