package cronos

import (
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	cronostypes "github.com/crypto-org-chain/cronos/v2/x/cronos/types"
)

type refreshBlocklistHandler func(blacklist []string)

type ProposalHandler struct {
	cronosKey                     *storetypes.KVStoreKey
	defaultProcessProposalHandler sdk.ProcessProposalHandler
	refreshBlocklistHandler       refreshBlocklistHandler
}

func NewProposalHandler(
	cronosKey *storetypes.KVStoreKey,
	mp mempool.Mempool,
	txVerifier baseapp.ProposalTxVerifier,
	refreshBlocklistHandler refreshBlocklistHandler,
) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mp, txVerifier)
	return &ProposalHandler{
		cronosKey:                     cronosKey,
		defaultProcessProposalHandler: defaultHandler.ProcessProposalHandler(),
		refreshBlocklistHandler:       refreshBlocklistHandler,
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req abci.RequestProcessProposal) abci.ResponseProcessProposal {
		store := ctx.KVStore(h.cronosKey)
		res := store.Get(cronostypes.KeyPrefixBlocklist)
		h.refreshBlocklistHandler(strings.Split(string(res), ","))
		return abci.ResponseProcessProposal{Status: abci.ResponseProcessProposal_ACCEPT}
	}
}
