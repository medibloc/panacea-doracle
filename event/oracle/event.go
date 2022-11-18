package oracle

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/medibloc/panacea-doracle/panacea"
	"github.com/tendermint/tendermint/light/provider"
)

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
