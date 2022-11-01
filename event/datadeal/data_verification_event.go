package datadeal

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/medibloc/panacea-doracle/validation"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var _ event.Event = (*DataVerificationEvent)(nil)

type DataVerificationEvent struct {
	reactor event.Reactor
}

func NewDataVerificationEvent(r event.Reactor) DataVerificationEvent {
	return DataVerificationEvent{r}
}

func (d DataVerificationEvent) GetEventType() string {
	return "message"
}

func (d DataVerificationEvent) GetEventAttributeKey() string {
	return "action"
}

func (d DataVerificationEvent) GetEventAttributeValue() string {
	return "'SellData'"
}

func (d DataVerificationEvent) EventHandler(event ctypes.ResultEvent) error {
	dealIDStr := event.Events[types.EventTypeDataVerificationVote+"."+types.AttributeKeyDealID][0]
	dataHash := event.Events[types.EventTypeDataVerificationVote+"."+types.AttributeKeyDataHash][0]

	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		return err
	}

	voteOption, err := d.verifyAndGetVoteOption(dealID, dataHash)
	if err != nil {
		if voteOption == oracletypes.VOTE_OPTION_UNSPECIFIED {
			return fmt.Errorf("can't vote due to error while verify. dealID(%d). dataHash(%s)", dealID, dataHash)
		} else {
			log.Infof("vote NO due to error while verifying. dealID(%d). dataHash(%s)", dealID, dataHash)
		}
	}

	msgVoteDataVerification, err := makeDataVerificationVote(
		d.reactor.OracleAcc().GetAddress(),
		dataHash,
		dealID,
		voteOption,
		d.reactor.OraclePrivKey().Serialize(),
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*d.reactor.QueryClient())

	txBytes, err := generateTxBytes(msgVoteDataVerification, d.reactor.OracleAcc().GetPrivKey(), d.reactor.Config(), txBuilder)
	if err != nil {
		return fmt.Errorf("generate tx failed. dealID(%d). dataHash(%s): %w", dealID, dataHash, err)
	}

	if err := broadcastTx(d.reactor.GRPCClient(), txBytes); err != nil {
		return fmt.Errorf("broadcast transaction failed. dealID(%d). dataHash(%s): %w", dealID, dataHash, err)
	}

	return nil
}

func (d DataVerificationEvent) decryptData(decryptedSharedKey, nonce, encryptedDataBz []byte) ([]byte, error) {
	decryptedData, err := crypto.DecryptWithAES256(decryptedSharedKey, nonce, encryptedDataBz)
	if err != nil {
		return nil, err
	}
	return decryptedData, nil
}

func (d DataVerificationEvent) compareDataHash(dataSale *types.DataSale, decryptedData []byte) bool {
	decryptedDataHash := sha256.Sum256(decryptedData)
	decryptedDataHashStr := hex.EncodeToString(decryptedDataHash[:])

	return decryptedDataHashStr == dataSale.DataHash
}

func (d DataVerificationEvent) convertSellerData(deal *types.Deal, dataSale *types.DataSale) ([]byte, error) {
	encryptedDataBz, err := d.reactor.Ipfs().Get(dataSale.VerifiableCid)
	if err != nil {
		log.Infof("failed to get data from IPFS: %v", err)
		return nil, nil
	}

	oraclePrivKey := d.reactor.OraclePrivKey()

	sellerAcc, err := d.reactor.QueryClient().GetAccount(dataSale.SellerAddress)
	if err != nil {
		log.Infof("failed to get account from panacea: %v", err)
		return nil, nil
	}

	sellerPubKeyBytes := sellerAcc.GetPubKey().Bytes()

	sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes, btcec.S256())
	if err != nil {
		log.Infof("failed to parsing sellerPubKey: %v", err)
		return nil, nil
	}

	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, sellerPubKey, crypto.KDFSHA256)

	decryptedData, err := d.decryptData(decryptSharedKey, deal.Nonce, encryptedDataBz)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

func (d DataVerificationEvent) verifyAndGetVoteOption(dealID uint64, dataHash string) (oracletypes.VoteOption, error) {
	deal, err := d.reactor.QueryClient().GetDeal(dealID)
	if err != nil {
		log.Infof("failed to find deal (%d)", dealID)
		return oracletypes.VOTE_OPTION_NO, nil
	}

	dataSale, err := d.reactor.QueryClient().GetDataSale(dataHash, dealID)
	if err != nil {
		log.Infof("failed to find dataSale (%s)", dataHash)
		return oracletypes.VOTE_OPTION_UNSPECIFIED, err
	}

	if dataSale.Status != types.DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD {
		return oracletypes.VOTE_OPTION_UNSPECIFIED, errors.New("dataSale's status is not DATA_SALE_STATUS_VERIFICATION_VOTING_PERIOD")
	}

	decryptedData, err := d.convertSellerData(deal, dataSale)
	if err != nil {
		log.Infof("failed to decrypt seller data, error (%v)", err)
		return oracletypes.VOTE_OPTION_NO, err
	}

	if !d.compareDataHash(dataSale, decryptedData) {
		log.Infof("invalid data hash")
		return oracletypes.VOTE_OPTION_NO, err
	}

	err = validation.ValidateJSONSchemata(decryptedData, deal.DataSchema)
	if err != nil {
		log.Infof("failed to verify data. error(%s)", err)
		return oracletypes.VOTE_OPTION_NO, err
	}

	return oracletypes.VOTE_OPTION_YES, nil
}

func makeDataVerificationVote(voterAddress, dataHash string, dealID uint64, voteOption oracletypes.VoteOption, oraclePrivKey []byte) (*types.MsgVoteDataVerification, error) {
	dataVerificationVote := &types.DataVerificationVote{
		VoterAddress: voterAddress,
		DealId:       dealID,
		DataHash:     dataHash,
		VoteOption:   voteOption,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledVerificationVote, err := dataVerificationVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledVerificationVote)
	if err != nil {
		return nil, err
	}

	msgVoteDataVerification := &types.MsgVoteDataVerification{
		DataVerificationVote: dataVerificationVote,
		Signature:            sig,
	}

	return msgVoteDataVerification, nil
}

// generateTxBytes generates transaction byte array.
// TODO: generateTxBytes function will be refactored.
func generateTxBytes(msgVoteDataVerification *types.MsgVoteDataVerification, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgVoteDataVerification)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// broadcastTx broadcast transaction to blockchain.
// TODO: broadcastTx function will be refactored.
func broadcastTx(grpcClient *panacea.GrpcClient, txBytes []byte) error {
	resp, err := grpcClient.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("data verification vote trasnsaction failed: %v", resp.TxResponse.RawLog)
	}

	log.Infof("MsgVoteDataVerification transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}
