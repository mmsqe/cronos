package cronos

import (
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	cronostypes "github.com/crypto-org-chain/cronos/v2/x/cronos/types"
)

type refreshBlocklistHandler func(blacklist []string)

type ProposalHandler struct {
	cdc                           codec.BinaryCodec
	cronosKey                     *storetypes.KVStoreKey
	stakingKey                    *storetypes.KVStoreKey
	defaultProcessProposalHandler sdk.ProcessProposalHandler
	refreshBlocklistHandler       refreshBlocklistHandler
}

func NewProposalHandler(
	cdc codec.BinaryCodec,
	cronosKey *storetypes.KVStoreKey,
	stakingKey *storetypes.KVStoreKey,
	mp mempool.Mempool,
	txVerifier baseapp.ProposalTxVerifier,
	refreshBlocklistHandler refreshBlocklistHandler,
) *ProposalHandler {
	defaultHandler := baseapp.NewDefaultProposalHandler(mp, txVerifier)
	return &ProposalHandler{
		cdc:                           cdc,
		cronosKey:                     cronosKey,
		stakingKey:                    stakingKey,
		defaultProcessProposalHandler: defaultHandler.ProcessProposalHandler(),
		refreshBlocklistHandler:       refreshBlocklistHandler,
	}
}

func (h *ProposalHandler) ProcessProposal() sdk.ProcessProposalHandler {
	return func(ctx sdk.Context, req abci.RequestProcessProposal) abci.ResponseProcessProposal {
		res := ctx.KVStore(h.cronosKey).Get(cronostypes.KeyPrefixBlocklist)
		h.refreshBlocklistHandler(strings.Split(string(res), ","))
		consAddr := sdk.ConsAddress(req.ProposerAddress)
		store := ctx.KVStore(h.stakingKey)
		opAddr := store.Get(stakingtypes.GetValidatorByConsAddrKey(consAddr))
		value := store.Get(stakingtypes.GetValidatorKey(sdk.ValAddress(opAddr)))
		if len(value) > 0 {
			validator := stakingtypes.MustUnmarshalValidator(h.cdc, value)
			fmt.Println("mm-validator", validator)
		}
		return h.defaultProcessProposalHandler(ctx, req)
	}
}
