package precompiles

import (
	"errors"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	cronosevents "github.com/crypto-org-chain/cronos/v2/x/cronos/events"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/events/bindings/cosmos/precompile/relayer"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/types"
)

var (
	irelayerABI                abi.ABI
	relayerContractAddress     = common.BytesToAddress([]byte{101})
	relayerGasRequiredByMethod = map[[4]byte]uint64{}
)

const (
	CreateClient                         = "createClient"
	UpdateClient                         = "updateClient"
	UpgradeClient                        = "upgradeClient"
	SubmitMisbehaviour                   = "submitMisbehaviour"
	ConnectionOpenInit                   = "connectionOpenInit"
	ConnectionOpenTry                    = "connectionOpenTry"
	ConnectionOpenAck                    = "connectionOpenAck"
	ConnectionOpenConfirm                = "connectionOpenConfirm"
	ChannelOpenInit                      = "channelOpenInit"
	ChannelOpenTry                       = "channelOpenTry"
	ChannelOpenAck                       = "channelOpenAck"
	ChannelOpenConfirm                   = "channelOpenConfirm"
	RecvPacket                           = "recvPacket"
	Acknowledgement                      = "acknowledgement"
	Timeout                              = "timeout"
	TimeoutOnClose                       = "timeoutOnClose"
	UpdateClientAndConnectionOpenInit    = "updateClientAndConnectionOpenInit"
	UpdateClientAndConnectionOpenTry     = "updateClientAndConnectionOpenTry"
	UpdateClientAndConnectionOpenAck     = "updateClientAndConnectionOpenAck"
	UpdateClientAndConnectionOpenConfirm = "updateClientAndConnectionOpenConfirm"
	UpdateClientAndChannelOpenInit       = "updateClientAndChannelOpenInit"
	UpdateClientAndChannelOpenTry        = "updateClientAndChannelOpenTry"
	UpdateClientAndChannelOpenAck        = "updateClientAndChannelOpenAck"
	UpdateClientAndChannelOpenConfirm    = "updateClientAndChannelOpenConfirm"
	UpdateClientAndRecvPacket            = "updateClientAndRecvPacket"
	UpdateClientAndAcknowledgement       = "updateClientAndAcknowledgement"
	UpdateClientAndTimeout               = "updateClientAndTimeout"
	UpdateClientAndTimeoutOnClose        = "updateClientAndTimeoutOnClose"
)

func init() {
	if err := irelayerABI.UnmarshalJSON([]byte(relayer.RelayerFunctionsMetaData.ABI)); err != nil {
		panic(err)
	}
	for methodName := range irelayerABI.Methods {
		var methodID [4]byte
		copy(methodID[:], irelayerABI.Methods[methodName].ID[:4])
		switch methodName {
		case CreateClient:
			relayerGasRequiredByMethod[methodID] = 200000
		case RecvPacket, Acknowledgement:
			relayerGasRequiredByMethod[methodID] = 250000
		case UpdateClient, UpgradeClient:
			relayerGasRequiredByMethod[methodID] = 400000
		default:
			relayerGasRequiredByMethod[methodID] = 100000
		}
	}
}

type RelayerContract struct {
	BaseContract

	cdc         codec.Codec
	ibcKeeper   types.IbcKeeper
	kvGasConfig storetypes.GasConfig
}

func NewRelayerContract(ibcKeeper types.IbcKeeper, cdc codec.Codec, kvGasConfig storetypes.GasConfig) vm.PrecompiledContract {
	return &RelayerContract{
		BaseContract: NewBaseContract(relayerContractAddress),
		ibcKeeper:    ibcKeeper,
		cdc:          cdc,
		kvGasConfig:  kvGasConfig,
	}
}

func (bc *RelayerContract) Address() common.Address {
	return relayerContractAddress
}

// RequiredGas calculates the contract gas use
func (bc *RelayerContract) RequiredGas(input []byte) uint64 {
	// base cost to prevent large input size
	baseCost := uint64(len(input)) * bc.kvGasConfig.WriteCostPerByte
	var methodID [4]byte
	copy(methodID[:], input[:4])
	requiredGas, ok := relayerGasRequiredByMethod[methodID]
	if ok {
		return requiredGas + baseCost
	}
	return baseCost
}

func (bc *RelayerContract) IsStateful() bool {
	return true
}

func (bc *RelayerContract) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	if readonly {
		return nil, errors.New("the method is not readonly")
	}
	// parse input
	methodID := contract.Input[:4]
	method, err := irelayerABI.MethodById(methodID)
	if err != nil {
		return nil, err
	}
	stateDB := evm.StateDB.(ExtStateDB)

	var res []byte
	precompileAddr := bc.Address()
	args, err := method.Inputs.Unpack(contract.Input[4:])
	if err != nil {
		return nil, errors.New("fail to unpack input arguments")
	}
	input := args[0].([]byte)
	converter := cronosevents.RelayerConvertEvent
	e := &Executor{
		cdc:       bc.cdc,
		stateDB:   stateDB,
		caller:    contract.CallerAddress,
		contract:  precompileAddr,
		input:     input,
		converter: converter,
	}
	if len(args) > 1 {
		e.input2 = args[1].([]byte)
	}
	switch method.Name {
	case CreateClient:
		res, err = exec(e, bc.ibcKeeper.CreateClient)
	case UpdateClient:
		res, err = exec(e, bc.ibcKeeper.UpdateClient)
	case UpgradeClient:
		res, err = exec(e, bc.ibcKeeper.UpgradeClient)
	case SubmitMisbehaviour:
		res, err = exec(e, bc.ibcKeeper.SubmitMisbehaviour)
	case ConnectionOpenInit:
		res, err = exec(e, bc.ibcKeeper.ConnectionOpenInit)
	case ConnectionOpenTry:
		res, err = exec(e, bc.ibcKeeper.ConnectionOpenTry)
	case ConnectionOpenAck:
		res, err = exec(e, bc.ibcKeeper.ConnectionOpenAck)
	case ConnectionOpenConfirm:
		res, err = exec(e, bc.ibcKeeper.ConnectionOpenConfirm)
	case ChannelOpenInit:
		res, err = exec(e, bc.ibcKeeper.ChannelOpenInit)
	case ChannelOpenTry:
		res, err = exec(e, bc.ibcKeeper.ChannelOpenTry)
	case ChannelOpenAck:
		res, err = exec(e, bc.ibcKeeper.ChannelOpenAck)
	case ChannelOpenConfirm:
		res, err = exec(e, bc.ibcKeeper.ChannelOpenConfirm)
	case RecvPacket:
		res, err = exec(e, bc.ibcKeeper.RecvPacket)
	case Acknowledgement:
		res, err = exec(e, bc.ibcKeeper.Acknowledgement)
	case Timeout:
		res, err = exec(e, bc.ibcKeeper.Timeout)
	case TimeoutOnClose:
		res, err = exec(e, bc.ibcKeeper.TimeoutOnClose)
	case UpdateClientAndConnectionOpenInit:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ConnectionOpenInit)
	case UpdateClientAndConnectionOpenTry:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ConnectionOpenTry)
	case UpdateClientAndConnectionOpenAck:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ConnectionOpenAck)
	case UpdateClientAndConnectionOpenConfirm:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ConnectionOpenConfirm)
	case UpdateClientAndChannelOpenInit:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ChannelOpenInit)
	case UpdateClientAndChannelOpenTry:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ChannelOpenTry)
	case UpdateClientAndChannelOpenAck:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ChannelOpenAck)
	case UpdateClientAndChannelOpenConfirm:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.ChannelOpenConfirm)
	case UpdateClientAndRecvPacket:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.RecvPacket)
	case UpdateClientAndAcknowledgement:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.Acknowledgement)
	case UpdateClientAndTimeout:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.Timeout)
	case UpdateClientAndTimeoutOnClose:
		res, err = execMultiple(e, bc.ibcKeeper.UpdateClient, bc.ibcKeeper.TimeoutOnClose)
	default:
		return nil, errors.New("unknown method")
	}
	return res, err
}
