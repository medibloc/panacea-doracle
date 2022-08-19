package panacea

import sdk "github.com/cosmos/cosmos-sdk/types"

const prefix = "panacea"

func GetAccAddressFromBech32(address string) (addr sdk.AccAddress, err error) {
	return sdk.GetFromBech32(address, prefix)
}
