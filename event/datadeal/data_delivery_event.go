package datadeal

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	datadealtypes "github.com/medibloc/panacea-core/v2/x/datadeal/types"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/config"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/event"
	"github.com/medibloc/panacea-doracle/panacea"
	log "github.com/sirupsen/logrus"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

type DataDeliveryVoteEvent struct {
	reactor event.Reactor
}

func NewDataDeliveryVoteEvent(s event.Reactor) DataDeliveryVoteEvent {
	return DataDeliveryVoteEvent{s}
}

func (e DataDeliveryVoteEvent) GetEventType() string {
	return "data_delivery"
}

func (e DataDeliveryVoteEvent) GetEventAttributeKey() string {
	return "vote_status"
}

func (e DataDeliveryVoteEvent) GetEventAttributeValue() string {
	return "'started'"
}

func (e DataDeliveryVoteEvent) EventHandler(event ctypes.ResultEvent) error {

	dealIDStr := event.Events[datadealtypes.EventTypeDataDeliveryVote+"."+datadealtypes.AttributeKeyDealID][0]
	dataHash := event.Events[datadealtypes.EventTypeDataDeliveryVote+"."+datadealtypes.AttributeKeyDataHash][0]

	dealID, err := strconv.ParseUint(dealIDStr, 10, 64)
	if err != nil {
		return err
	}

	oraclePrivKey := e.reactor.OraclePrivKey()

	voteOption, deliveredCid, err := e.verifyAndGetVoteOption(dealID, dataHash, oraclePrivKey)
	if err != nil {
		log.Errorf("error while verify: %v", err)
		if voteOption == oracletypes.VOTE_OPTION_UNSPECIFIED {
			return err
		}
	}

	msgVoteDataDelivery, err := e.makeDataDeliveryVote(
		e.reactor.OracleAcc().GetAddress(),
		dataHash,
		deliveredCid,
		dealID,
		voteOption,
		oraclePrivKey.Serialize(),
	)
	if err != nil {
		return err
	}

	txBuilder := panacea.NewTxBuilder(*e.reactor.QueryClient())

	txBytes, err := e.generateTxBytes(msgVoteDataDelivery, e.reactor.OracleAcc().GetPrivKey(), e.reactor.Config(), txBuilder)
	if err != nil {
		return err
	}

	if err := e.broadcastTx(e.reactor.GRPCClient(), txBytes); err != nil {
		return err
	}

	return nil

}

func (e DataDeliveryVoteEvent) verifyAndGetVoteOption(dealID uint64, dataHash string, oraclePrivKey *btcec.PrivateKey) (oracletypes.VoteOption, string, error) {

	dataSale, err := e.reactor.QueryClient().GetDataSale(dealID, dataHash)
	if err != nil {
		if errors.Is(err, datadealtypes.ErrDataSaleNotFound) {
			return oracletypes.VOTE_OPTION_NO, "", err
		} else {
			return oracletypes.VOTE_OPTION_UNSPECIFIED, "", err
		}
	}

	if dataSale.Status != datadealtypes.DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD {
		return oracletypes.VOTE_OPTION_UNSPECIFIED, "", errors.New("datasale status is not DATA_SALE_STATUS_DELIVERY_VOTING_PERIOD")
	}

	if len(dataSale.VerifiableCid) == 0 {
		return oracletypes.VOTE_OPTION_NO, "", errors.New("there is no verifiavleCid")
	}

	deal, err := e.reactor.QueryClient().GetDeal(dataSale.DealId)
	if err != nil {
		if errors.Is(err, datadealtypes.ErrDealNotFound) {
			return oracletypes.VOTE_OPTION_NO, "", err
		} else {
			return oracletypes.VOTE_OPTION_UNSPECIFIED, "", err
		}
	}

	deliveredCID, err := e.convertBuyerDataAndAddToIpfs(deal, dataSale, oraclePrivKey)
	if err != nil {
		return oracletypes.VOTE_OPTION_NO, "", fmt.Errorf("error while make deliveredCid: %v", err)
	}

	return oracletypes.VOTE_OPTION_YES, deliveredCID, nil

}

func (e DataDeliveryVoteEvent) convertBuyerDataAndAddToIpfs(deal *datadealtypes.Deal, dataSale *datadealtypes.DataSale, oraclePrivKey *btcec.PrivateKey) (string, error) {
	// get encrypted data from ipfs
	encryptedDataBz, err := e.reactor.Ipfs().Get(dataSale.VerifiableCid)
	if err != nil {
		return "", err
	}

	// get shared key oraclePrivKey + sellerPublicKey
	sellerAcc, err := e.reactor.QueryClient().GetAccount(dataSale.SellerAddress)
	if err != nil {
		return "", err
	}
	sellerPubKeyBytes := sellerAcc.GetPubKey().Bytes()

	sellerPubKey, err := btcec.ParsePubKey(sellerPubKeyBytes, btcec.S256())
	if err != nil {
		return "", err
	}
	decryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, sellerPubKey, crypto.KDFSHA256)

	// decrypt data
	decryptedData, err := crypto.DecryptWithAES256(decryptSharedKey, deal.Nonce, encryptedDataBz)
	if err != nil {
		return "", err
	}

	// get oraclePrivateKey & buyerPublicKey and make shared key
	buyerAccount, err := e.reactor.QueryClient().GetAccount(deal.BuyerAddress)
	if err != nil {
		return "", err
	}

	buyerPubKeyBytes := buyerAccount.GetPubKey().Bytes()
	buyerPubKey, err := btcec.ParsePubKey(buyerPubKeyBytes, btcec.S256())
	if err != nil {
		return "", err
	}
	encryptSharedKey := crypto.DeriveSharedKey(oraclePrivKey, buyerPubKey, crypto.KDFSHA256)

	// encrypt data
	encryptDataWithBuyerKey, err := crypto.EncryptWithAES256(encryptSharedKey, deal.Nonce, decryptedData)
	if err != nil {
		return "", err
	}

	// ipfs.add (decrypted data) & get CID
	deliveredCid, err := e.reactor.Ipfs().Add(encryptDataWithBuyerKey)
	if err != nil {
		return "", err
	}

	return deliveredCid, nil
}

func (e DataDeliveryVoteEvent) makeDataDeliveryVote(voterAddress, dataHash, deliveredCid string, dealID uint64, voteOption oracletypes.VoteOption, oraclePrivKey []byte) (*datadealtypes.MsgVoteDataDelivery, error) {

	dataDeliveryVote := &datadealtypes.DataDeliveryVote{
		VoterAddress: voterAddress,
		DealId:       dealID,
		DataHash:     dataHash,
		DeliveredCid: deliveredCid,
		VoteOption:   voteOption,
	}

	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledDataDeliveryVote, err := dataDeliveryVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledDataDeliveryVote)
	if err != nil {
		return nil, err
	}

	msgVoteDataDelivery := &datadealtypes.MsgVoteDataDelivery{
		DataDeliveryVote: dataDeliveryVote,
		Signature:        sig,
	}

	return msgVoteDataDelivery, nil
}

func (e DataDeliveryVoteEvent) generateTxBytes(msgVoteDataDelivery *datadealtypes.MsgVoteDataDelivery, privKey cryptotypes.PrivKey, conf *config.Config, txBuilder *panacea.TxBuilder) ([]byte, error) {
	defaultFeeAmount, _ := sdk.ParseCoinsNormalized(conf.Panacea.DefaultFeeAmount)
	txBytes, err := txBuilder.GenerateSignedTxBytes(privKey, conf.Panacea.DefaultGasLimit, defaultFeeAmount, msgVoteDataDelivery)
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}

// broadcastTx broadcast transaction to blockchain.
func (e DataDeliveryVoteEvent) broadcastTx(grpcClient *panacea.GrpcClient, txBytes []byte) error {
	resp, err := grpcClient.BroadcastTx(txBytes)
	if err != nil {
		return err
	}

	if resp.TxResponse.Code != 0 {
		return fmt.Errorf("data delivery vote trasnsaction failed: %v", resp.TxResponse.RawLog)
	}

	log.Infof("MsgVoteDataDelivery transaction succeed. height(%v), hash(%s)", resp.TxResponse.Height, resp.TxResponse.TxHash)

	return nil
}
