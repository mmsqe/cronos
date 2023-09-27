// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package icacallback

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ICACallbackMetaData contains all meta data concerning the ICACallback contract.
var ICACallbackMetaData = &bind.MetaData{
	ABI: "[{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"internalType\":\"uint64\",\"name\":\"seq\",\"type\":\"uint64\"},{\"indexed\":false,\"internalType\":\"bytes\",\"name\":\"ack\",\"type\":\"bytes\"}],\"name\":\"OnPacketResult\",\"type\":\"event\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"name\":\"acknowledgement\",\"outputs\":[{\"internalType\":\"bytes\",\"name\":\"\",\"type\":\"bytes\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"name\":\"getLastAckSeq\",\"outputs\":[{\"internalType\":\"uint256\",\"name\":\"\",\"type\":\"uint256\"}],\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"seq\",\"type\":\"uint64\"},{\"internalType\":\"address\",\"name\":\"packetSenderAddress\",\"type\":\"address\"},{\"internalType\":\"bytes\",\"name\":\"ack\",\"type\":\"bytes\"}],\"name\":\"onPacketResult\",\"outputs\":[{\"internalType\":\"bool\",\"name\":\"\",\"type\":\"bool\"}],\"stateMutability\":\"payable\",\"type\":\"function\"},{\"inputs\":[{\"internalType\":\"uint64\",\"name\":\"seq\",\"type\":\"uint64\"},{\"internalType\":\"bytes\",\"name\":\"ack\",\"type\":\"bytes\"}],\"name\":\"setLastAckSeq\",\"outputs\":[{\"internalType\":\"uint64\",\"name\":\"\",\"type\":\"uint64\"}],\"stateMutability\":\"nonpayable\",\"type\":\"function\"}]",
	Bin: "0x608060405234801561000f575f80fd5b50610add8061001d5f395ff3fe60806040526004361061003e575f3560e01c806315400f96146100425780637a78e6a514610072578063968bcba3146100ae578063dcae8cfd146100ea575b5f80fd5b61005c60048036038101906100579190610498565b610114565b6040516100699190610523565b60405180910390f35b34801561007d575f80fd5b506100986004803603810190610093919061053c565b610211565b6040516100a591906105a8565b60405180910390f35b3480156100b9575f80fd5b506100d460048036038101906100cf91906105c1565b6102d8565b6040516100e19190610676565b60405180910390f35b3480156100f5575f80fd5b506100fe610373565b60405161010b91906106ae565b60405180910390f35b5f3073ffffffffffffffffffffffffffffffffffffffff168473ffffffffffffffffffffffffffffffffffffffff1614610183576040517f08c379a000000000000000000000000000000000000000000000000000000000815260040161017a90610721565b60405180910390fd5b828260025f8867ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f2091826101b7929190610970565b508282600191826101c9929190610970565b507f8e0c6cb5698eba8240951fde76f9e06a0844d4285c0e56f4cedf1415d03703fc8584846040516101fd93929190610a77565b60405180910390a160019050949350505050565b5f835f806101000a81548167ffffffffffffffff021916908367ffffffffffffffff16021790555082826001918261024a929190610970565b50828260025f8767ffffffffffffffff1667ffffffffffffffff1681526020019081526020015f20918261027f929190610970565b507f8e0c6cb5698eba8240951fde76f9e06a0844d4285c0e56f4cedf1415d03703fc8484846040516102b393929190610a77565b60405180910390a15f8054906101000a900467ffffffffffffffff1690509392505050565b6002602052805f5260405f205f9150905080546102f4906107a3565b80601f0160208091040260200160405190810160405280929190818152602001828054610320906107a3565b801561036b5780601f106103425761010080835404028352916020019161036b565b820191905f5260205f20905b81548152906001019060200180831161034e57829003601f168201915b505050505081565b5f805f9054906101000a900467ffffffffffffffff1667ffffffffffffffff16905090565b5f80fd5b5f80fd5b5f67ffffffffffffffff82169050919050565b6103bc816103a0565b81146103c6575f80fd5b50565b5f813590506103d7816103b3565b92915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f610406826103dd565b9050919050565b610416816103fc565b8114610420575f80fd5b50565b5f813590506104318161040d565b92915050565b5f80fd5b5f80fd5b5f80fd5b5f8083601f84011261045857610457610437565b5b8235905067ffffffffffffffff8111156104755761047461043b565b5b6020830191508360018202830111156104915761049061043f565b5b9250929050565b5f805f80606085870312156104b0576104af610398565b5b5f6104bd878288016103c9565b94505060206104ce87828801610423565b935050604085013567ffffffffffffffff8111156104ef576104ee61039c565b5b6104fb87828801610443565b925092505092959194509250565b5f8115159050919050565b61051d81610509565b82525050565b5f6020820190506105365f830184610514565b92915050565b5f805f6040848603121561055357610552610398565b5b5f610560868287016103c9565b935050602084013567ffffffffffffffff8111156105815761058061039c565b5b61058d86828701610443565b92509250509250925092565b6105a2816103a0565b82525050565b5f6020820190506105bb5f830184610599565b92915050565b5f602082840312156105d6576105d5610398565b5b5f6105e3848285016103c9565b91505092915050565b5f81519050919050565b5f82825260208201905092915050565b5f5b83811015610623578082015181840152602081019050610608565b5f8484015250505050565b5f601f19601f8301169050919050565b5f610648826105ec565b61065281856105f6565b9350610662818560208601610606565b61066b8161062e565b840191505092915050565b5f6020820190508181035f83015261068e818461063e565b905092915050565b5f819050919050565b6106a881610696565b82525050565b5f6020820190506106c15f83018461069f565b92915050565b5f82825260208201905092915050565b7f646966666572656e742073656e646572000000000000000000000000000000005f82015250565b5f61070b6010836106c7565b9150610716826106d7565b602082019050919050565b5f6020820190508181035f830152610738816106ff565b9050919050565b5f82905092915050565b7f4e487b71000000000000000000000000000000000000000000000000000000005f52604160045260245ffd5b7f4e487b71000000000000000000000000000000000000000000000000000000005f52602260045260245ffd5b5f60028204905060018216806107ba57607f821691505b6020821081036107cd576107cc610776565b5b50919050565b5f819050815f5260205f209050919050565b5f6020601f8301049050919050565b5f82821b905092915050565b5f6008830261082f7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff826107f4565b61083986836107f4565b95508019841693508086168417925050509392505050565b5f819050919050565b5f61087461086f61086a84610696565b610851565b610696565b9050919050565b5f819050919050565b61088d8361085a565b6108a16108998261087b565b848454610800565b825550505050565b5f90565b6108b56108a9565b6108c0818484610884565b505050565b5b818110156108e3576108d85f826108ad565b6001810190506108c6565b5050565b601f821115610928576108f9816107d3565b610902846107e5565b81016020851015610911578190505b61092561091d856107e5565b8301826108c5565b50505b505050565b5f82821c905092915050565b5f6109485f198460080261092d565b1980831691505092915050565b5f6109608383610939565b9150826002028217905092915050565b61097a838361073f565b67ffffffffffffffff81111561099357610992610749565b5b61099d82546107a3565b6109a88282856108e7565b5f601f8311600181146109d5575f84156109c3578287013590505b6109cd8582610955565b865550610a34565b601f1984166109e3866107d3565b5f5b82811015610a0a578489013582556001820191506020850194506020810190506109e5565b86831015610a275784890135610a23601f891682610939565b8355505b6001600288020188555050505b50505050505050565b828183375f83830152505050565b5f610a5683856105f6565b9350610a63838584610a3d565b610a6c8361062e565b840190509392505050565b5f604082019050610a8a5f830186610599565b8181036020830152610a9d818486610a4b565b905094935050505056fea2646970667358221220ffc8c02b719773be7267ae5fb9def95900b4747c7fe5bca26416c7f511e7d03c64736f6c63430008150033",
}

// ICACallbackABI is the input ABI used to generate the binding from.
// Deprecated: Use ICACallbackMetaData.ABI instead.
var ICACallbackABI = ICACallbackMetaData.ABI

// ICACallbackBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ICACallbackMetaData.Bin instead.
var ICACallbackBin = ICACallbackMetaData.Bin

// DeployICACallback deploys a new Ethereum contract, binding an instance of ICACallback to it.
func DeployICACallback(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *ICACallback, error) {
	parsed, err := ICACallbackMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ICACallbackBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ICACallback{ICACallbackCaller: ICACallbackCaller{contract: contract}, ICACallbackTransactor: ICACallbackTransactor{contract: contract}, ICACallbackFilterer: ICACallbackFilterer{contract: contract}}, nil
}

// ICACallback is an auto generated Go binding around an Ethereum contract.
type ICACallback struct {
	ICACallbackCaller     // Read-only binding to the contract
	ICACallbackTransactor // Write-only binding to the contract
	ICACallbackFilterer   // Log filterer for contract events
}

// ICACallbackCaller is an auto generated read-only Go binding around an Ethereum contract.
type ICACallbackCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ICACallbackTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ICACallbackTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ICACallbackFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ICACallbackFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ICACallbackSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ICACallbackSession struct {
	Contract     *ICACallback      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ICACallbackCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ICACallbackCallerSession struct {
	Contract *ICACallbackCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// ICACallbackTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ICACallbackTransactorSession struct {
	Contract     *ICACallbackTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// ICACallbackRaw is an auto generated low-level Go binding around an Ethereum contract.
type ICACallbackRaw struct {
	Contract *ICACallback // Generic contract binding to access the raw methods on
}

// ICACallbackCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ICACallbackCallerRaw struct {
	Contract *ICACallbackCaller // Generic read-only contract binding to access the raw methods on
}

// ICACallbackTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ICACallbackTransactorRaw struct {
	Contract *ICACallbackTransactor // Generic write-only contract binding to access the raw methods on
}

// NewICACallback creates a new instance of ICACallback, bound to a specific deployed contract.
func NewICACallback(address common.Address, backend bind.ContractBackend) (*ICACallback, error) {
	contract, err := bindICACallback(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ICACallback{ICACallbackCaller: ICACallbackCaller{contract: contract}, ICACallbackTransactor: ICACallbackTransactor{contract: contract}, ICACallbackFilterer: ICACallbackFilterer{contract: contract}}, nil
}

// NewICACallbackCaller creates a new read-only instance of ICACallback, bound to a specific deployed contract.
func NewICACallbackCaller(address common.Address, caller bind.ContractCaller) (*ICACallbackCaller, error) {
	contract, err := bindICACallback(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ICACallbackCaller{contract: contract}, nil
}

// NewICACallbackTransactor creates a new write-only instance of ICACallback, bound to a specific deployed contract.
func NewICACallbackTransactor(address common.Address, transactor bind.ContractTransactor) (*ICACallbackTransactor, error) {
	contract, err := bindICACallback(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ICACallbackTransactor{contract: contract}, nil
}

// NewICACallbackFilterer creates a new log filterer instance of ICACallback, bound to a specific deployed contract.
func NewICACallbackFilterer(address common.Address, filterer bind.ContractFilterer) (*ICACallbackFilterer, error) {
	contract, err := bindICACallback(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ICACallbackFilterer{contract: contract}, nil
}

// bindICACallback binds a generic wrapper to an already deployed contract.
func bindICACallback(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ICACallbackABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ICACallback *ICACallbackRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ICACallback.Contract.ICACallbackCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ICACallback *ICACallbackRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ICACallback.Contract.ICACallbackTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ICACallback *ICACallbackRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ICACallback.Contract.ICACallbackTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ICACallback *ICACallbackCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ICACallback.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ICACallback *ICACallbackTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ICACallback.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ICACallback *ICACallbackTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ICACallback.Contract.contract.Transact(opts, method, params...)
}

// Acknowledgement is a free data retrieval call binding the contract method 0x968bcba3.
//
// Solidity: function acknowledgement(uint64 ) view returns(bytes)
func (_ICACallback *ICACallbackCaller) Acknowledgement(opts *bind.CallOpts, arg0 uint64) ([]byte, error) {
	var out []interface{}
	err := _ICACallback.contract.Call(opts, &out, "acknowledgement", arg0)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// Acknowledgement is a free data retrieval call binding the contract method 0x968bcba3.
//
// Solidity: function acknowledgement(uint64 ) view returns(bytes)
func (_ICACallback *ICACallbackSession) Acknowledgement(arg0 uint64) ([]byte, error) {
	return _ICACallback.Contract.Acknowledgement(&_ICACallback.CallOpts, arg0)
}

// Acknowledgement is a free data retrieval call binding the contract method 0x968bcba3.
//
// Solidity: function acknowledgement(uint64 ) view returns(bytes)
func (_ICACallback *ICACallbackCallerSession) Acknowledgement(arg0 uint64) ([]byte, error) {
	return _ICACallback.Contract.Acknowledgement(&_ICACallback.CallOpts, arg0)
}

// GetLastAckSeq is a free data retrieval call binding the contract method 0xdcae8cfd.
//
// Solidity: function getLastAckSeq() view returns(uint256)
func (_ICACallback *ICACallbackCaller) GetLastAckSeq(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ICACallback.contract.Call(opts, &out, "getLastAckSeq")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetLastAckSeq is a free data retrieval call binding the contract method 0xdcae8cfd.
//
// Solidity: function getLastAckSeq() view returns(uint256)
func (_ICACallback *ICACallbackSession) GetLastAckSeq() (*big.Int, error) {
	return _ICACallback.Contract.GetLastAckSeq(&_ICACallback.CallOpts)
}

// GetLastAckSeq is a free data retrieval call binding the contract method 0xdcae8cfd.
//
// Solidity: function getLastAckSeq() view returns(uint256)
func (_ICACallback *ICACallbackCallerSession) GetLastAckSeq() (*big.Int, error) {
	return _ICACallback.Contract.GetLastAckSeq(&_ICACallback.CallOpts)
}

// OnPacketResult is a paid mutator transaction binding the contract method 0x15400f96.
//
// Solidity: function onPacketResult(uint64 seq, address packetSenderAddress, bytes ack) payable returns(bool)
func (_ICACallback *ICACallbackTransactor) OnPacketResult(opts *bind.TransactOpts, seq uint64, packetSenderAddress common.Address, ack []byte) (*types.Transaction, error) {
	return _ICACallback.contract.Transact(opts, "onPacketResult", seq, packetSenderAddress, ack)
}

// OnPacketResult is a paid mutator transaction binding the contract method 0x15400f96.
//
// Solidity: function onPacketResult(uint64 seq, address packetSenderAddress, bytes ack) payable returns(bool)
func (_ICACallback *ICACallbackSession) OnPacketResult(seq uint64, packetSenderAddress common.Address, ack []byte) (*types.Transaction, error) {
	return _ICACallback.Contract.OnPacketResult(&_ICACallback.TransactOpts, seq, packetSenderAddress, ack)
}

// OnPacketResult is a paid mutator transaction binding the contract method 0x15400f96.
//
// Solidity: function onPacketResult(uint64 seq, address packetSenderAddress, bytes ack) payable returns(bool)
func (_ICACallback *ICACallbackTransactorSession) OnPacketResult(seq uint64, packetSenderAddress common.Address, ack []byte) (*types.Transaction, error) {
	return _ICACallback.Contract.OnPacketResult(&_ICACallback.TransactOpts, seq, packetSenderAddress, ack)
}

// SetLastAckSeq is a paid mutator transaction binding the contract method 0x7a78e6a5.
//
// Solidity: function setLastAckSeq(uint64 seq, bytes ack) returns(uint64)
func (_ICACallback *ICACallbackTransactor) SetLastAckSeq(opts *bind.TransactOpts, seq uint64, ack []byte) (*types.Transaction, error) {
	return _ICACallback.contract.Transact(opts, "setLastAckSeq", seq, ack)
}

// SetLastAckSeq is a paid mutator transaction binding the contract method 0x7a78e6a5.
//
// Solidity: function setLastAckSeq(uint64 seq, bytes ack) returns(uint64)
func (_ICACallback *ICACallbackSession) SetLastAckSeq(seq uint64, ack []byte) (*types.Transaction, error) {
	return _ICACallback.Contract.SetLastAckSeq(&_ICACallback.TransactOpts, seq, ack)
}

// SetLastAckSeq is a paid mutator transaction binding the contract method 0x7a78e6a5.
//
// Solidity: function setLastAckSeq(uint64 seq, bytes ack) returns(uint64)
func (_ICACallback *ICACallbackTransactorSession) SetLastAckSeq(seq uint64, ack []byte) (*types.Transaction, error) {
	return _ICACallback.Contract.SetLastAckSeq(&_ICACallback.TransactOpts, seq, ack)
}

// ICACallbackOnPacketResultIterator is returned from FilterOnPacketResult and is used to iterate over the raw logs and unpacked data for OnPacketResult events raised by the ICACallback contract.
type ICACallbackOnPacketResultIterator struct {
	Event *ICACallbackOnPacketResult // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ICACallbackOnPacketResultIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ICACallbackOnPacketResult)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ICACallbackOnPacketResult)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ICACallbackOnPacketResultIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ICACallbackOnPacketResultIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ICACallbackOnPacketResult represents a OnPacketResult event raised by the ICACallback contract.
type ICACallbackOnPacketResult struct {
	Seq uint64
	Ack []byte
	Raw types.Log // Blockchain specific contextual infos
}

// FilterOnPacketResult is a free log retrieval operation binding the contract event 0x8e0c6cb5698eba8240951fde76f9e06a0844d4285c0e56f4cedf1415d03703fc.
//
// Solidity: event OnPacketResult(uint64 seq, bytes ack)
func (_ICACallback *ICACallbackFilterer) FilterOnPacketResult(opts *bind.FilterOpts) (*ICACallbackOnPacketResultIterator, error) {

	logs, sub, err := _ICACallback.contract.FilterLogs(opts, "OnPacketResult")
	if err != nil {
		return nil, err
	}
	return &ICACallbackOnPacketResultIterator{contract: _ICACallback.contract, event: "OnPacketResult", logs: logs, sub: sub}, nil
}

// WatchOnPacketResult is a free log subscription operation binding the contract event 0x8e0c6cb5698eba8240951fde76f9e06a0844d4285c0e56f4cedf1415d03703fc.
//
// Solidity: event OnPacketResult(uint64 seq, bytes ack)
func (_ICACallback *ICACallbackFilterer) WatchOnPacketResult(opts *bind.WatchOpts, sink chan<- *ICACallbackOnPacketResult) (event.Subscription, error) {

	logs, sub, err := _ICACallback.contract.WatchLogs(opts, "OnPacketResult")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ICACallbackOnPacketResult)
				if err := _ICACallback.contract.UnpackLog(event, "OnPacketResult", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseOnPacketResult is a log parse operation binding the contract event 0x8e0c6cb5698eba8240951fde76f9e06a0844d4285c0e56f4cedf1415d03703fc.
//
// Solidity: event OnPacketResult(uint64 seq, bytes ack)
func (_ICACallback *ICACallbackFilterer) ParseOnPacketResult(log types.Log) (*ICACallbackOnPacketResult, error) {
	event := new(ICACallbackOnPacketResult)
	if err := _ICACallback.contract.UnpackLog(event, "OnPacketResult", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
