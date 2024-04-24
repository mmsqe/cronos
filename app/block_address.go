package app

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BlockAddressesDecorator block addresses from sending transactions
type BlockAddressesDecorator struct {
	blockedMapGetter func() map[string]struct{}
}

func NewBlockAddressesDecorator(blockedMapGetter func() map[string]struct{}) *BlockAddressesDecorator {
	return &BlockAddressesDecorator{
		blockedMapGetter: blockedMapGetter,
	}
}

func (bad *BlockAddressesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	blockedMap := bad.blockedMapGetter()
	if ctx.IsCheckTx() {
		for _, msg := range tx.GetMsgs() {
			for _, signer := range msg.GetSigners() {
				if _, ok := blockedMap[string(signer)]; ok {
					return ctx, fmt.Errorf("signer is blocked: %s", signer.String())
				}
			}
		}
	}
	return next(ctx, tx, simulate)
}
