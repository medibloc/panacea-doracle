package oracle

import (
	"testing"

	"github.com/btcsuite/btcd/btcec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	oracletypes "github.com/medibloc/panacea-core/v2/x/oracle/types"
	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/stretchr/testify/require"
)

func TestMakeOracleRegistrationVoteYes(t *testing.T) {
	address, err := bech32.ConvertAndEncode(panacea.HRP, secp256k1.GenPrivKey().PubKey().Address().Bytes())
	require.NoError(t, err)

	uniqueID := "uniqueID"
	voteOption := oracletypes.VOTE_OPTION_YES
	oraclePrivKey, _ := btcec.NewPrivateKey(btcec.S256())
	nodePrivKey, _ := btcec.NewPrivateKey(btcec.S256())
	nodePubKey := nodePrivKey.PubKey().SerializeCompressed()
	nonce := []byte("123412341234")
	orv, err := makeMsgVoteOracleRegistration(uniqueID, address, address, voteOption, oraclePrivKey.Serialize(), nodePubKey, nonce)

	require.NoError(t, err)
	require.Equal(t, uniqueID, orv.OracleRegistrationVote.UniqueId)
	require.Equal(t, address, orv.OracleRegistrationVote.VoterAddress)
	require.Equal(t, address, orv.OracleRegistrationVote.VotingTargetAddress)
	require.Equal(t, voteOption, orv.OracleRegistrationVote.VoteOption)
	require.NotNil(t, orv.OracleRegistrationVote.EncryptedOraclePrivKey)
}

func TestMakeOracleRegistrationVoteNo(t *testing.T) {
	address, err := bech32.ConvertAndEncode(panacea.HRP, secp256k1.GenPrivKey().PubKey().Address().Bytes())
	require.NoError(t, err)

	uniqueID := "uniqueID"
	voteOption := oracletypes.VOTE_OPTION_NO
	oraclePrivKey, _ := btcec.NewPrivateKey(btcec.S256())
	nodePrivKey, _ := btcec.NewPrivateKey(btcec.S256())
	nodePubKey := nodePrivKey.PubKey().SerializeCompressed()
	nonce := []byte("123412341234")
	orv, err := makeMsgVoteOracleRegistration(uniqueID, address, address, voteOption, oraclePrivKey.Serialize(), nodePubKey, nonce)

	require.NoError(t, err)
	require.Equal(t, uniqueID, orv.OracleRegistrationVote.UniqueId)
	require.Equal(t, address, orv.OracleRegistrationVote.VoterAddress)
	require.Equal(t, address, orv.OracleRegistrationVote.VotingTargetAddress)
	require.Equal(t, voteOption, orv.OracleRegistrationVote.VoteOption)
	require.Nil(t, orv.OracleRegistrationVote.EncryptedOraclePrivKey)
}
