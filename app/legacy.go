package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	icaauthtypes "github.com/crypto-org-chain/cronos/v2/app/legacy/icaauth/types"
)

func RegisterLegacyCodec(cdc *codec.LegacyAmino) {
	icaauthtypes.RegisterCodec(cdc)
}

func RegisterLegacyInterfaces(registry cdctypes.InterfaceRegistry) {
	icaauthtypes.RegisterInterfaces(registry)
}
