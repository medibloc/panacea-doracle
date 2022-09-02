package panacea

import (
	"context"
	"errors"
	"fmt"
	ics23 "github.com/confio/ics23/go"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/ibc-go/v2/modules/core/23-commitment/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	sgxdb "github.com/medibloc/panacea-doracle/store/sgxleveldb"
	log "github.com/sirupsen/logrus"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/light"
	"github.com/tendermint/tendermint/light/provider"
	tmhttp "github.com/tendermint/tendermint/light/provider/http"
	dbs "github.com/tendermint/tendermint/light/store/db"
	"github.com/tendermint/tendermint/rpc/client"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"
	tmtypes "github.com/tendermint/tendermint/types"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var DbDir string

type TrustedBlockInfo struct {
	TrustedBlockHeight int64
	TrustedBlockHash   []byte
}

type QueryClient struct {
	RpcClient         *rpchttp.HTTP
	LightClient       *light.Client
	interfaceRegistry codectypes.InterfaceRegistry
	sgxLevelDB        *sgxdb.SgxLevelDB
	mutex             *sync.Mutex
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	DbDir = filepath.Join(userHomeDir, ".doracle", "data")
}

// NewQueryClient set QueryClient with rpcClient & and returns, if successful,
// a QueryClient that can be used to add query function.
func NewQueryClient(ctx context.Context, config *config.Config, info TrustedBlockInfo) (*QueryClient, error) {
	lcMutex := sync.Mutex{}
	chainID := config.Panacea.ChainID
	rpcClient, err := rpchttp.New(config.Panacea.RpcAddr, "/websocket")
	if err != nil {
		return nil, err
	}

	trustOptions := light.TrustOptions{
		Period: 2 * 365 * 24 * time.Hour,
		Height: info.TrustedBlockHeight,
		Hash:   info.TrustedBlockHash,
	}

	pv, err := tmhttp.New(chainID, config.Panacea.PrimaryAddr)
	if err != nil {
		return nil, err
	}

	var pvs []provider.Provider
	witnessAddrs := strings.Split(config.Panacea.WitnessesAddr, ",")
	for _, witnessAddr := range witnessAddrs {
		witness, err := tmhttp.New(chainID, witnessAddr)
		if err != nil {
			return nil, err
		}
		pvs = append(pvs, witness)
	}
	db, err := sgxdb.NewSgxLevelDB("light-client-db", DbDir)
	if err != nil {
		return nil, err
	}

	store := dbs.New(db, chainID)

	lc, err := light.NewClient(
		ctx,
		chainID,
		trustOptions,
		pv,
		pvs,
		store,
		light.SkippingVerification(light.DefaultTrustLevel),
		light.Logger(tmlog.TestingLogger()),
	)
	if err != nil {
		return nil, err
	}

	// call refresh every minute
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			if err := refresh(ctx, lc, trustOptions.Period, &lcMutex); err != nil {
				log.Errorf("light client refresh error: %v", err)
			}
		}
	}()

	return &QueryClient{
		RpcClient:         rpcClient,
		LightClient:       lc,
		interfaceRegistry: makeInterfaceRegistry(),
		sgxLevelDB:        db,
		mutex:             &lcMutex,
	}, nil
}

func (q QueryClient) safeUpdateLightClient(ctx context.Context) (*tmtypes.LightBlock, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.LightClient.Update(ctx, time.Now())
}

func (q QueryClient) safeVerifyLightBlockAtHeight(ctx context.Context, height int64) (*tmtypes.LightBlock, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	return q.LightClient.VerifyLightBlockAtHeight(ctx, height, time.Now())
}

// refresh update light block, if the last light block has been updated more than trustPeriod * 2/3 ago.
func refresh(ctx context.Context, lc *light.Client, trustPeriod time.Duration, m *sync.Mutex) error {
	log.Info("check latest light block")
	lastBlockHeight, err := lc.LastTrustedHeight()
	if err != nil {
		return err
	}
	lastBlock, err := lc.TrustedLightBlock(lastBlockHeight)
	if err != nil {
		return err
	}
	lastBlockTime := lastBlock.Time
	currentTime := time.Now()
	timeDiff := currentTime.Sub(lastBlockTime)
	if timeDiff > trustPeriod*2/3 {
		log.Info("update latest light block")
		m.Lock()
		defer m.Unlock()
		if _, err := lc.Update(ctx, time.Now()); err != nil {
			return err
		}
	}

	return nil
}

// GetStoreData get data from panacea with storeKey and key, then verify queried data with light client and merkle proof.
// the returned data type is ResponseQuery.value ([]byte), so recommend to convert to expected type
func (q QueryClient) GetStoreData(ctx context.Context, storeKey string, key []byte) ([]byte, error) {
	var queryHeight int64

	// get recent light block
	// if the latest block has already been updated, get LastTrustedHeight
	trustedBlock, err := q.safeUpdateLightClient(ctx)
	if err != nil {
		return nil, err
	}
	if trustedBlock == nil {
		queryHeight, err = q.LightClient.LastTrustedHeight()
		if err != nil {
			return nil, err
		}
	} else {
		queryHeight = trustedBlock.Height
	}

	//set queryOption prove to true
	option := client.ABCIQueryOptions{
		Prove:  true,
		Height: queryHeight,
	}
	// query to kv store with proof option
	result, err := q.RpcClient.ABCIQueryWithOptions(ctx, fmt.Sprintf("/store/%s/key", storeKey), key, option)
	if err != nil {
		return nil, err
	}

	// for merkle proof, the apphash of the next block is needed.
	// requests the next block every second, and generates an error after more than 12sec
	var nextTrustedBlock *tmtypes.LightBlock
	i := 0
	for {
		nextTrustedBlock, err = q.safeVerifyLightBlockAtHeight(ctx, queryHeight+1)
		if errors.Is(err, provider.ErrHeightTooHigh) {
			time.Sleep(1 * time.Second)
			i++
		} else if err != nil {
			return nil, err
		} else {
			break
		}
		if i > 12 {
			return nil, fmt.Errorf("can not get nextTrustedBlock")
		}
	}

	// verify query result with merkle proof & trusted block info
	proofOps := result.Response.ProofOps
	merkleProof, err := types.ConvertProofs(proofOps)
	if err != nil {
		return nil, err
	}

	sdkSpecs := []*ics23.ProofSpec{ics23.IavlSpec, ics23.TendermintSpec}
	merkleRootKey := types.NewMerkleRoot(nextTrustedBlock.AppHash.Bytes())

	merklePath := types.NewMerklePath(storeKey, string(key))
	err = merkleProof.VerifyMembership(sdkSpecs, merkleRootKey, merklePath, result.Response.Value)
	if err != nil {
		return nil, err
	}

	return result.Response.Value, nil
}

func (q QueryClient) Close() error {
	err := q.sgxLevelDB.Close()
	if err != nil {
		return err
	}

	q.RpcClient.OnStop()

	return nil
}

// Below are examples of query function that use GetStoreData function to verify queried result.
// Need to set storeKey and key inside the query function, and change type to expected type.

// GetAccount returns account from address.
func (q QueryClient) GetAccount(address string) (authtypes.AccountI, error) {
	acc, err := GetAccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	key := authtypes.AddressStoreKey(acc)
	bz, err := q.GetStoreData(context.Background(), authtypes.StoreKey, key)
	if err != nil {
		return nil, err
	}

	var accountAny codectypes.Any
	err = accountAny.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	var account authtypes.AccountI
	err = q.interfaceRegistry.UnpackAny(&accountAny, &account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func (q QueryClient) GetOracleRegistration(oracleAddr, uniqueID string) (*oracletypes.OracleRegistration, error) {

	acc, err := GetAccAddressFromBech32(oracleAddr)
	if err != nil {
		return nil, err
	}

	key := oracletypes.GetOracleRegistrationKey(uniqueID, acc)

	bz, err := q.GetStoreData(context.Background(), oracletypes.StoreKey, key)
	if err != nil {
		return nil, err
	}
	var oracleRegistration oracletypes.OracleRegistration
	err = oracleRegistration.Unmarshal(bz)
	if err != nil {
		return nil, err
	}

	return &oracleRegistration, nil
}
