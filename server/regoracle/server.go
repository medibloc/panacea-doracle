package regoracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/edgelesssys/ego/enclave"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/server/service"
	"github.com/medibloc/panacea-doracle/sgx"
	"github.com/medibloc/panacea-doracle/types"
	log "github.com/sirupsen/logrus"
	tos "github.com/tendermint/tendermint/libs/os"
	"os"
	"path/filepath"
)

var nodePrivKeyPath string

type Server struct {
	*service.Service
}

func NewServer(svc *service.Service) *Server {
	return &Server{
		Service: svc,
	}
}

func (srv *Server) Run() error {
	// if node key exists, continue the process of registration with the existing key
	if tos.FileExists(nodePrivKeyPath) {
		return srv.continueRegistration()
	}

	// if there is no node private key, create a new node key and request OracleRegistration
	return srv.createRegistration()
}

func (srv *Server) Close() error {
	return nil
}

// continueRegistration continues the oracle registration process using the existing node key
func (srv *Server) continueRegistration() error {
	// get existing node key
	nodePrivKeyBz, err := sgx.UnsealFromFile(nodePrivKeyPath)
	if err != nil {
		return err
	}
	_, nodePubKey := crypto.PrivKeyFromBytes(nodePrivKeyBz)

	// get unique ID
	selfEnclaveInfo, err := sgx.GetSelfEnclaveInfo()
	if err != nil {
		return err
	}

	uniqueID := base64.StdEncoding.EncodeToString(selfEnclaveInfo.UniqueID)

	// get oracle registration from Panacea.
	// TODO: Should be replaced with light client query for data verification
	oracleRegistration, err := srv.PanaceaClient.GetOracleRegistration(srv.OracleAccount.GetAddress(), uniqueID)
	if err != nil {
		return err
	}

	// check if the same node key is used for oracle registration
	if !bytes.Equal(oracleRegistration.NodePubKey, nodePubKey.SerializeCompressed()) {
		return errors.New("the existing node key is different from the one used in oracle registration. if you want to re-request RegisterOracle, delete the existing node_priv_key.sealed file and rerun register-oracle cmd")
	}

	// handle differently depending on the status of oracle registration
	switch oracleRegistration.Status {
	case oracletypes.ORACLE_REGISTRATION_STATUS_VOTING_PERIOD:
		// in voting period
		// subscribe an event for OracleRegistrationCompleted
	case oracletypes.ORACLE_REGISTRATION_STATUS_PASSED:
		// oracle registration vote already passed
		// check if oracle private key exists or not
		// if exists, return nil. no need to do register-oracle
		// else, get encryptedOraclePrivKey from Panacea and decrypt and SealToFile it
	case oracletypes.ORACLE_REGISTRATION_STATUS_REJECTED:
		// oracle registration vote rejected
		return errors.New("the request for oracle registration is rejected. please delete the existing node key and retry with a new one")
	default:
		return errors.New("invalid oracle registration status")
	}

	return nil
}

// createRegistration generates a random node key pair and request for oracle registration to Panacea
func (srv *Server) createRegistration() error {
	//generate node key and its remote report
	nodePubKey, nodePubKeyRemoteReport, err := generateNodeKey()
	if err != nil {
		log.Errorf("failed to generate node key pair: %v", err)
		return err
	}

	report, _ := enclave.VerifyRemoteReport(nodePubKeyRemoteReport)
	uniqueIDStr := base64.StdEncoding.EncodeToString(report.UniqueID)

	// sign and broadcast to Panacea
	msgRegisterOracle := oracletypes.NewMsgRegisterOracle(uniqueIDStr, srv.OracleAccount.GetAddress(), nodePubKey, nodePubKeyRemoteReport, srv.TrustedHeight, srv.TrustedHash)

	txBuilder := panacea.NewTxBuilder(srv.PanaceaClient)

	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(srv.Conf.Panacea.DefaultFeeAmount)

	txBytes, err := txBuilder.GenerateSignedTxBytes(srv.OracleAccount.GetPrivKey(), srv.Conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgRegisterOracle)
	if err != nil {
		log.Errorf("failed to generate signed Tx bytes: %v", err)
		return err
	}

	response, err := srv.PanaceaClient.BroadcastTx(txBytes)
	if err != nil {
		log.Errorf("failed to broadcast transaction: %v", err)
		return err
	}

	// if tx fails,
	if response.TxResponse.Code != 0 {
		log.Errorf("transaction for registration of oracle failed: %v", response.TxResponse.Logs)
		return fmt.Errorf("transaction failed")
	}

	// if tx success, let's wait for vote. maybe use WaitGroup?
	// subscriber := event.NewSubscriber(event.RegisterOracleCompleted)
	// subscribe event and handle it
	// defer subscriber.Close()
	return nil
}

// generateNodeKey generates random node key and its remote report
// And the generated private key is sealed and stored
func generateNodeKey() ([]byte, []byte, error) {
	nodePrivKey, err := crypto.NewPrivKey()
	if err != nil {
		return nil, nil, err
	}

	if err := sgx.SealToFile(nodePrivKey.Serialize(), nodePrivKeyPath); err != nil {
		return nil, nil, err
	}

	nodePubKey := nodePrivKey.PubKey().SerializeCompressed()
	nodePubKeyHash := sha256.Sum256(nodePubKey)
	nodeKeyRemoteReport, err := sgx.GenerateRemoteReport(nodePubKeyHash[:])
	if err != nil {
		return nil, nil, err
	}

	return nodePubKey, nodeKeyRemoteReport, nil
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	nodePrivKeyPath = filepath.Join(userHomeDir, types.DefaultDoracleDir, types.DefaultNodePrivKeyName)
}
