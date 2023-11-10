package precompiles

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/cometbft/cometbft/libs/log"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	cronosevents "github.com/crypto-org-chain/cronos/v2/x/cronos/events"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/events/bindings/cosmos/precompile/ica"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/events/bindings/cosmos/precompile/icacallback"
	cronoseventstypes "github.com/crypto-org-chain/cronos/v2/x/cronos/events/types"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/types"

	icaauthtypes "github.com/crypto-org-chain/cronos/v2/x/icaauth/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	RegisterAccountMethodName = "registerAccount"
	QueryAccountMethodName    = "queryAccount"
	SubmitMsgsMethodName      = "submitMsgs"
)

var (
	icaABI                 abi.ABI
	icaCallbackABI         abi.ABI
	icaContractAddress     = common.BytesToAddress([]byte{102})
	icaMethodNamedByMethod = map[[4]byte]string{}
	icaGasRequiredByMethod = map[[4]byte]uint64{}
)

func init() {
	if err := icaABI.UnmarshalJSON([]byte(ica.ICAModuleMetaData.ABI)); err != nil {
		panic(err)
	}
	if err := icaCallbackABI.UnmarshalJSON([]byte(icacallback.ICACallbackMetaData.ABI)); err != nil {
		panic(err)
	}

	for methodName := range icaABI.Methods {
		var methodID [4]byte
		copy(methodID[:], icaABI.Methods[methodName].ID[:4])
		switch methodName {
		case RegisterAccountMethodName:
			icaGasRequiredByMethod[methodID] = 231455
		case QueryAccountMethodName:
			icaGasRequiredByMethod[methodID] = 100000
		case SubmitMsgsMethodName:
			icaGasRequiredByMethod[methodID] = 83086
		default:
			icaGasRequiredByMethod[methodID] = 0
		}
		icaMethodNamedByMethod[methodID] = methodName
	}
}

func OnPacketResultCallback(args ...interface{}) ([]byte, error) {
	return icaCallbackABI.Pack("onPacketResultCallback", args...)
}

type IcaContract struct {
	BaseContract

	cdc           codec.Codec
	icaauthKeeper types.Icaauthkeeper
	cronosKeeper  types.CronosKeeper
	logger        log.Logger
}

func NewIcaContract(icaauthKeeper types.Icaauthkeeper, cronosKeeper types.CronosKeeper, cdc codec.Codec, logger log.Logger) vm.PrecompiledContract {
	return &IcaContract{
		BaseContract:  NewBaseContract(icaContractAddress),
		cdc:           cdc,
		icaauthKeeper: icaauthKeeper,
		cronosKeeper:  cronosKeeper,
		logger:        logger.With("precompiles", "ica"),
	}
}

func (ic *IcaContract) Address() common.Address {
	return icaContractAddress
}

// RequiredGas calculates the contract gas use
// `max(0, len(input) * DefaultTxSizeCostPerByte + requiredGasTable[methodPrefix] - intrinsicGas)`
func (ic *IcaContract) RequiredGas(input []byte) (gas uint64) {
	intrinsicGas, err := core.IntrinsicGas(input, nil, false, true, true)
	if err != nil {
		return 0
	}
	var methodID [4]byte
	copy(methodID[:], input[:4])
	requiredGas, ok := icaGasRequiredByMethod[methodID]
	if !ok {
		requiredGas = 0
	}
	// base cost to prevent large input size
	baseCost := uint64(len(input)) * authtypes.DefaultTxSizeCostPerByte
	defer func() {
		method := icaMethodNamedByMethod[methodID]
		ic.logger.Debug("required", "gas", gas, "method", method, "len", len(input), "intrinsic", intrinsicGas)
	}()
	total := requiredGas + baseCost
	if total < intrinsicGas {
		return 0
	}
	return total - intrinsicGas
}

func (ic *IcaContract) IsStateful() bool {
	return true
}

func (ic *IcaContract) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	// parse input
	methodID := contract.Input[:4]
	method, err := icaABI.MethodById(methodID)
	if err != nil {
		return nil, err
	}
	stateDB := evm.StateDB.(ExtStateDB)
	precompileAddr := ic.Address()
	caller := contract.CallerAddress
	owner := sdk.AccAddress(caller.Bytes()).String()
	converter := cronosevents.IcaConvertEvent
	var execErr error
	switch method.Name {
	case RegisterAccountMethodName:
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := method.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		connectionID := args[0].(string)
		version := args[1].(string)
		execErr = stateDB.ExecuteNativeAction(precompileAddr, converter, func(ctx sdk.Context) error {
			_, err := ic.icaauthKeeper.RegisterAccount(ctx, &icaauthtypes.MsgRegisterAccount{
				Owner:        owner,
				ConnectionId: connectionID,
				Version:      version,
			})
			return err
		})
		if execErr != nil {
			return nil, execErr
		}
		return method.Outputs.Pack(true)
	case QueryAccountMethodName:
		args, err := method.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		connectionID := args[0].(string)
		account := args[1].(common.Address)
		owner := sdk.AccAddress(account.Bytes()).String()
		icaAddress := ""
		response, err := ic.icaauthKeeper.InterchainAccountAddress(
			stateDB.CacheContext(),
			&icaauthtypes.QueryInterchainAccountAddressRequest{
				Owner:        owner,
				ConnectionId: connectionID,
			})
		if err != nil {
			return nil, err
		}
		if response != nil {
			icaAddress = response.InterchainAccountAddress
		}
		return method.Outputs.Pack(icaAddress)
	case SubmitMsgsMethodName:
		if readonly {
			return nil, errors.New("the method is not readonly")
		}
		args, err := method.Inputs.Unpack(contract.Input[4:])
		if err != nil {
			return nil, errors.New("fail to unpack input arguments")
		}
		connectionID := args[0].(string)
		data := args[1].([]byte)
		timeout := args[2].(*big.Int)
		icaMsgData := icatypes.InterchainAccountPacketData{
			Type: icatypes.EXECUTE_TX,
			Data: data,
			Memo: fmt.Sprintf(`{"src_callback": {"address": "%s"}}`, caller.String()),
		}
		timeoutDuration := time.Duration(timeout.Uint64())
		seq := uint64(0)
		execErr = stateDB.ExecuteNativeAction(precompileAddr, converter, func(ctx sdk.Context) error {
			response, err := ic.icaauthKeeper.SubmitTxWithArgs(
				ctx,
				owner,
				connectionID,
				timeoutDuration,
				icaMsgData,
			)
			if err == nil && response != nil {
				seq = response.Sequence
				ctx.EventManager().EmitEvents(sdk.Events{
					sdk.NewEvent(
						cronoseventstypes.EventTypeSubmitMsgsResult,
						sdk.NewAttribute(cronoseventstypes.AttributeKeySeq, fmt.Sprintf("%d", response.Sequence)),
					),
				})
			}
			return err
		})
		if execErr != nil {
			return nil, execErr
		}
		return method.Outputs.Pack(seq)
	default:
		return nil, errors.New("unknown method")
	}
}
