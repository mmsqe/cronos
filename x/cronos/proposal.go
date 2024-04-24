package cronos

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	cronostypes "github.com/crypto-org-chain/cronos/v2/x/cronos/types"
)

type ProposalHandler struct {
	cronosKey                     *storetypes.KVStoreKey
	defaultProcessProposalHandler sdk.ProcessProposalHandler
}

func NewProposalHandler(
	cronosKey *storetypes.KVStoreKey,
	mp mempool.Mempool,
	txVerifier baseapp.ProposalTxVerifier,
) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mp, txVerifier)
	return &ProposalHandler{
		cronosKey:                     cronosKey,
		defaultProcessProposalHandler: defaultHandler.ProcessProposalHandler(),
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req abci.RequestProcessProposal) abci.ResponseProcessProposal {
		store := ctx.KVStore(h.cronosKey)
		res := store.Get(cronostypes.KeyPrefixBlocklist)
		fmt.Println("mm-res", string(res))
		return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
	}
}
