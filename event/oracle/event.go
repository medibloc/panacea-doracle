package oracle

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/crypto"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/tendermint/tendermint/light/provider"
)

func makeMsgVoteOracleRegistration(uniqueID, voterUniqueID, voterAddr, votingTargetAddr string, voteOption oracletypes.VoteOption, oraclePrivKey, nodePubKey, nonce []byte) (*oracletypes.MsgVoteOracleRegistration, error) {
	if oracletypes.VOTE_OPTION_YES == voteOption {
		privKey, _ := crypto.PrivKeyFromBytes(oraclePrivKey)
		pubKey, err := btcec.ParsePubKey(nodePubKey, btcec.S256())
		if err != nil {
			return nil, err
		}

		shareKey := crypto.DeriveSharedKey(privKey, pubKey, crypto.KDFSHA256)
		encryptedOraclePrivKey, err := crypto.EncryptWithAES256(shareKey, nonce, oraclePrivKey)
		if err != nil {
			return nil, err
		}

		registrationVote := &oracletypes.OracleRegistrationVote{
			UniqueId:               uniqueID,
			VoterUniqueId:          voterUniqueID,
			VoterAddress:           voterAddr,
			VotingTargetAddress:    votingTargetAddr,
			VoteOption:             oracletypes.VOTE_OPTION_YES,
			EncryptedOraclePrivKey: encryptedOraclePrivKey,
		}

		return makeMsgVoteOracleRegistrationWithSignature(registrationVote, oraclePrivKey)
	} else {
		return makeMsgVoteOracleRegistrationVoteTypeNo(
			uniqueID,
			voterUniqueID,
			voterAddr,
			votingTargetAddr,
			oraclePrivKey)
	}
}

func makeMsgVoteOracleRegistrationVoteTypeNo(uniqueID, voterUniqueID, voterAddr, votingTargetAddr string, oraclePrivKey []byte) (*oracletypes.MsgVoteOracleRegistration, error) {
	registrationVote := &oracletypes.OracleRegistrationVote{
		UniqueId:            uniqueID,
		VoterUniqueId:       voterUniqueID,
		VoterAddress:        voterAddr,
		VotingTargetAddress: votingTargetAddr,
		VoteOption:          oracletypes.VOTE_OPTION_NO,
	}

	return makeMsgVoteOracleRegistrationWithSignature(registrationVote, oraclePrivKey)

}

func makeMsgVoteOracleRegistrationWithSignature(registrationVote *oracletypes.OracleRegistrationVote, oraclePrivKey []byte) (*oracletypes.MsgVoteOracleRegistration, error) {
	key := secp256k1.PrivKey{
		Key: oraclePrivKey,
	}

	marshaledRegistrationVote, err := registrationVote.Marshal()
	if err != nil {
		return nil, err
	}

	sig, err := key.Sign(marshaledRegistrationVote)
	if err != nil {
		return nil, err
	}

	msgVoteOracleRegistration := &oracletypes.MsgVoteOracleRegistration{
		OracleRegistrationVote: registrationVote,
		Signature:              sig,
	}

	return msgVoteOracleRegistration, nil
}

func verifyTrustedBlockInfo(queryClient *panacea.QueryClient, height int64, blockHash []byte) error {
	block, err := queryClient.GetLightBlock(height)
	if err != nil {
		switch err {
		case provider.ErrLightBlockNotFound, provider.ErrHeightTooHigh:
			return fmt.Errorf("not found light block. %w", err)
		default:
			return err
		}
	}

	if !bytes.Equal(block.Hash().Bytes(), blockHash) {
		return fmt.Errorf("failed to verify trusted block information. height(%v), expected block hash(%s), got block hash(%s)",
			height,
			hex.EncodeToString(block.Hash().Bytes()),
			hex.EncodeToString(blockHash),
		)
	}

	return nil
}
