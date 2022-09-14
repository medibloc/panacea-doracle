package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type blockResp struct {
	BlockID struct {
		Hash string `json:"hash"`
	} `json:"block_id"`
	Block struct {
		Header struct {
			Height string `json:"height"`
		} `json:"header"`
	} `json:"block"`
}

func QueryLatestBlock(endpoint string) (string, int64, error) {
	resp, err := http.Get(fmt.Sprintf("%s/blocks/latest", endpoint))
	if err != nil {
		return "", 0, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", 0, fmt.Errorf("query returned non-200 status: %d", resp.StatusCode)
	}

	var result blockResp
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, fmt.Errorf("failed to read resp body: %w", err)
	}

	if result.BlockID.Hash == "" { // no block has been produced yet
		return "", 0, nil
	}

	height, err := strconv.ParseInt(result.Block.Header.Height, 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse num string: %w", err)
	}

	return result.BlockID.Hash, height, nil
}
