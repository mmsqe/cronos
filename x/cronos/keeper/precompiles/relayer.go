package precompiles

import (
	"errors"
	"fmt"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/gogoproto/proto"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	cronosevents "github.com/crypto-org-chain/cronos/v2/x/cronos/events"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/events/bindings/cosmos/precompile/relayer"
	"github.com/crypto-org-chain/cronos/v2/x/cronos/types"
	ethermint "github.com/evmos/ethermint/types"
)

var (
	irelayerABI                abi.ABI
	relayerContractAddress     = common.BytesToAddress([]byte{101})
	relayerMethodNamedByMethod = map[[4]byte]string{}
)

const (
	CreateClient          = "createClient"
	UpdateClient          = "updateClient"
	UpgradeClient         = "upgradeClient"
	SubmitMisbehaviour    = "submitMisbehaviour"
	ConnectionOpenInit    = "connectionOpenInit"
	ConnectionOpenTry     = "connectionOpenTry"
	ConnectionOpenAck     = "connectionOpenAck"
	ConnectionOpenConfirm = "connectionOpenConfirm"
	ChannelOpenInit       = "channelOpenInit"
	ChannelOpenTry        = "channelOpenTry"
	ChannelOpenAck        = "channelOpenAck"
	ChannelOpenConfirm    = "channelOpenConfirm"
	ChannelCloseInit      = "channelCloseInit"
	ChannelCloseConfirm   = "channelCloseConfirm"
	RecvPacket            = "recvPacket"
	Acknowledgement       = "acknowledgement"
	Timeout               = "timeout"
	TimeoutOnClose        = "timeoutOnClose"
)

func init() {
	if err := irelayerABI.UnmarshalJSON([]byte(relayer.RelayerFunctionsMetaData.ABI)); err != nil {
		panic(err)
	}
}

type SimulateFn func(txBytes []byte) (sdk.GasInfo, *sdk.Result, error)

type RelayerContract struct {
	BaseContract

	ctx         sdk.Context
	cdc         codec.Codec
	txConfig    client.TxConfig
	ibcKeeper   types.IbcKeeper
	simulate    SimulateFn
	logger      log.Logger
	isHomestead bool
	isIstanbul  bool
	isShanghai  bool
}

func NewRelayerContract(ctx sdk.Context, txConfig client.TxConfig, ibcKeeper types.IbcKeeper, simulate SimulateFn, cdc codec.Codec, rules params.Rules, logger log.Logger) vm.PrecompiledContract {
	return &RelayerContract{
		BaseContract: NewBaseContract(relayerContractAddress),
		ctx:          ctx,
		txConfig:     txConfig,
		ibcKeeper:    ibcKeeper,
		simulate:     simulate,
		cdc:          cdc,
		isHomestead:  rules.IsHomestead,
		isIstanbul:   rules.IsIstanbul,
		isShanghai:   rules.IsShanghai,
		logger:       logger.With("precompiles", "relayer"),
	}
}

func (bc *RelayerContract) Address() common.Address {
	return relayerContractAddress
}

func unpackInput(input []byte) ([][]byte, abi.Type, error) {
	t, err := abi.NewType("bytes[]", "", nil)
	if err != nil {
		return nil, t, err
	}
	var inputs [][]byte
	args, err := abi.Arguments{{
		Type: t,
	}}.Unpack(input)
	if err != nil {
		inputs = [][]byte{input}
	} else {
		inputs = append(inputs, args[0].([][]byte)...)
	}
	return inputs, t, nil
}

// RequiredGas calculates the contract gas use
// `max(0, len(input) * DefaultTxSizeCostPerByte + requiredGasTable[methodPrefix] - intrinsicGas)`
func (bc *RelayerContract) RequiredGas(input []byte) (finalGas uint64) {
	baseCost := uint64(15500)
	inputs, _, err := unpackInput(input)
	if err != nil {
		panic(err)
	}

	var msgs []proto.Message
	var signer sdk.AccAddress
	intrinsicGasTotal := uint64(0)
	for _, input := range inputs {
		var methodID [4]byte
		copy(methodID[:], input[:4])
		method, err := irelayerABI.MethodById(methodID[:])
		if err != nil {
			panic(err)
		}
		args, err := method.Inputs.Unpack(input[4:])
		if err != nil {
			panic(err)
		}
		i := args[0].([]byte)

		e := &Executor{
			cdc:   bc.cdc,
			input: i,
		}
		var msg NativeMessage
		switch method.Name {
		case CreateClient:
			msg, err = extractMsg[clienttypes.MsgCreateClient](e)
		case UpdateClient:
			msg, err = extractMsg[clienttypes.MsgUpdateClient](e)
		case UpgradeClient:
			msg, err = extractMsg[clienttypes.MsgUpgradeClient](e)
		case ConnectionOpenInit:
			msg, err = extractMsg[connectiontypes.MsgConnectionOpenInit](e)
		case ConnectionOpenTry:
			msg, err = extractMsg[connectiontypes.MsgConnectionOpenTry](e)
		case ConnectionOpenAck:
			msg, err = extractMsg[connectiontypes.MsgConnectionOpenAck](e)
		case ConnectionOpenConfirm:
			msg, err = extractMsg[connectiontypes.MsgConnectionOpenConfirm](e)
		case ChannelOpenInit:
			msg, err = extractMsg[channeltypes.MsgChannelOpenInit](e)
		case ChannelOpenTry:
			msg, err = extractMsg[channeltypes.MsgChannelOpenTry](e)
		case ChannelOpenAck:
			msg, err = extractMsg[channeltypes.MsgChannelOpenAck](e)
		case ChannelOpenConfirm:
			msg, err = extractMsg[channeltypes.MsgChannelOpenConfirm](e)
		case ChannelCloseInit:
			msg, err = extractMsg[channeltypes.MsgChannelCloseInit](e)
		case ChannelCloseConfirm:
			msg, err = extractMsg[channeltypes.MsgChannelCloseConfirm](e)
		case RecvPacket:
			msg, err = extractMsg[channeltypes.MsgRecvPacket](e)
		case Acknowledgement:
			msg, err = extractMsg[channeltypes.MsgAcknowledgement](e)
		case Timeout:
			msg, err = extractMsg[channeltypes.MsgTimeout](e)
		case TimeoutOnClose:
			msg, err = extractMsg[channeltypes.MsgTimeoutOnClose](e)
		default:
			panic(fmt.Errorf("unknown method: %s", method.Name))
		}
		if err != nil {
			panic(err)
		}

		msgs = append(msgs, msg)
		if signer == nil {
			signers := msg.GetSigners()
			if len(signers) != 1 {
				panic(errors.New("don't support multi-signers message"))
			}
			signer = signers[0]
		}
		intrinsicGas, _ := core.IntrinsicGas(input, nil, false, bc.isHomestead, bc.isIstanbul, bc.isShanghai)
		methodName := relayerMethodNamedByMethod[methodID]
		fmt.Println("mm-required", methodName, "intrinsic", intrinsicGas)
		intrinsicGasTotal += intrinsicGas
	}
	i, err := bc.buildSimTx(msgs...)
	if err != nil {
		panic(err)
	}
	g, _, err := bc.simulate(i)
	if err != nil {
		panic(err)
	}
	return g.GasUsed + baseCost - intrinsicGasTotal
}

// buildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be built.
func (bc *RelayerContract) buildSimTx(msgs ...sdk.Msg) ([]byte, error) {
	txf := tx.Factory{}
	txf = txf.
		WithChainID(bc.ctx.ChainID()).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT).
		WithTxConfig(bc.txConfig)

	max, ok := sdkmath.NewIntFromString("1000000")
	if !ok {
		return nil, fmt.Errorf("invalid opt value")
	}
	extensionOption := ethermint.ExtensionOptionDynamicFeeTx{
		MaxPriorityPrice: max,
	}
	extBytes, err := extensionOption.Marshal()
	if err != nil {
		return nil, err
	}
	// TODO: config extension option
	extOpts := []*cdctypes.Any{{
		TypeUrl: "/ethermint.types.v1.ExtensionOptionDynamicFeeTx",
		Value:   extBytes,
	}}
	txf = txf.WithExtensionOptions(extOpts...)
	txb, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		Data: &signing.SingleSignatureData{
			SignMode: txf.SignMode(),
		},
		Sequence: 1,
	}
	if err = txb.SetSignatures(sig); err != nil {
		return nil, err
	}
	txEncoder := bc.txConfig.TxEncoder()
	return txEncoder(txb.GetTx())
}

func (bc *RelayerContract) Run(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error) {
	if readonly {
		return nil, errors.New("the method is not readonly")
	}
	inputs, t, err := unpackInput(contract.Input)
	if err != nil {
		return nil, err
	}

	var responses [][]byte
	for _, input := range inputs {
		if len(input) < 4 {
			return nil, errors.New("input too short")
		}
		// parse input
		methodID := input[:4]
		method, err := irelayerABI.MethodById(methodID)
		if err != nil {
			return nil, err
		}
		stateDB := evm.StateDB.(ExtStateDB)

		var res []byte
		precompileAddr := bc.Address()
		args, err := method.Inputs.Unpack(input[4:])
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
		switch method.Name {
		case CreateClient:
			res, err = exec(e, bc.ibcKeeper.CreateClient)
		case UpdateClient:
			res, err = exec(e, bc.ibcKeeper.UpdateClient)
		case UpgradeClient:
			res, err = exec(e, bc.ibcKeeper.UpgradeClient)
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
		case ChannelCloseInit:
			res, err = exec(e, bc.ibcKeeper.ChannelCloseInit)
		case ChannelCloseConfirm:
			res, err = exec(e, bc.ibcKeeper.ChannelCloseConfirm)
		case RecvPacket:
			res, err = exec(e, bc.ibcKeeper.RecvPacket)
		case Acknowledgement:
			res, err = exec(e, bc.ibcKeeper.Acknowledgement)
		case Timeout:
			res, err = exec(e, bc.ibcKeeper.Timeout)
		case TimeoutOnClose:
			res, err = exec(e, bc.ibcKeeper.TimeoutOnClose)
		default:
			return nil, fmt.Errorf("unknown method: %s", method.Name)
		}
		if err != nil {
			return nil, err
		}
		responses = append(responses, res)
	}
	return abi.Arguments{{
		Type: t,
	}}.Pack(responses)
}
