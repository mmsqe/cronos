package types

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var (
	Ten       = big.NewInt(10)
	TenPowTen = Ten.Exp(Ten, Ten, nil)
)

const (
	ibcDenomPrefix    = "ibc/"
	ibcDenomLen       = len(ibcDenomPrefix) + 64
	cronosDenomPrefix = "cronos0x"
	cronosDenomLen    = len(cronosDenomPrefix) + 40
)

// IsValidIBCDenom returns true if denom is a valid ibc denom
func IsValidIBCDenom(denom string) bool {
	return len(denom) == ibcDenomLen && strings.HasPrefix(denom, ibcDenomPrefix)
}

// IsValidCronosDenom returns true if denom is a valid cronos denom
func IsValidCronosDenom(denom string) bool {
	return len(denom) == cronosDenomLen && strings.HasPrefix(denom, cronosDenomPrefix)
}

// IsSourceCoin returns true if denom is a coin originated from cronos
func IsSourceCoin(denom string) bool {
	return IsValidCronosDenom(denom)
}

// IsValidCoinDenom returns true if it's ok it is a valid coin denom
func IsValidCoinDenom(denom string) bool {
	return IsValidIBCDenom(denom) || IsValidCronosDenom(denom)
}

// GetContractAddressFromDenom get the contract address from the coin denom
func GetContractAddressFromDenom(denom string) (string, error) {
	contractAddress := ""
	if strings.HasPrefix(denom, cronosDenomPrefix) {
		contractAddress = denom[6:]
	}
	if !common.IsHexAddress(contractAddress) {
		return "", fmt.Errorf("invalid contract address (%s)", contractAddress)
	}
	return contractAddress, nil
}
