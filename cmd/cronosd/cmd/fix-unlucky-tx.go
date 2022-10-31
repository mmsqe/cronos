package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/crypto-org-chain/cronos/app"
	"github.com/spf13/cobra"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/eth/tracers"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/params"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcfg "github.com/tendermint/tendermint/config"
	tmnode "github.com/tendermint/tendermint/node"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/state/indexer/sink/psql"
	"github.com/tendermint/tendermint/state/txindex"
	"github.com/tendermint/tendermint/state/txindex/kv"
	"github.com/tendermint/tendermint/state/txindex/null"
	tmstore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/client"
	rpctypes "github.com/evmos/ethermint/rpc/types"
	ethermint "github.com/evmos/ethermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ethtypes "github.com/ethereum/go-ethereum/core/types"

	"github.com/evmos/ethermint/x/evm/statedb"
	"github.com/evmos/ethermint/x/evm/types"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
)

const (
	FlagStartBlock  = "start-block"
	FlagEndBlock    = "end-block"
	FlagConcurrency = "concurrency"
)

// FixUnluckyTxCmd update the tx execution result of false-failed tx in tendermint db
func FixUnluckyTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "fix-unlucky-tx",
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (returnErr error) {
			now := time.Now()
			defer func() {
				fmt.Println("total time: ", time.Since(now))
			}()
			ctx := server.GetServerContextFromCmd(cmd)
			chainID, err := cmd.Flags().GetString(flags.FlagChainID)
			if err != nil {
				return err
			}
			tmDB, err := openTMDB(ctx.Config, chainID)
			if err != nil {
				return err
			}
			state, err := tmDB.stateStore.Load()
			if err != nil {
				return err
			}

			appDB, err := openAppDB(ctx.Config.RootDir)
			if err != nil {
				return err
			}
			defer func() {
				if err := appDB.Close(); err != nil {
					ctx.Logger.With("error", err).Error("error closing db")
				}
			}()

			idxDB, err := openIndexerDB(ctx.Config.RootDir)
			if err != nil {
				return err
			}
			defer func() {
				if err := idxDB.Close(); err != nil {
					ctx.Logger.With("error", err).Error("error closing idxDB")
				}
			}()
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			// idxer := indexer.NewKVIndexer(idxDB, ctx.Logger.With("module", "evmindex"), clientCtx)

			encCfg := app.MakeEncodingConfig()
			appCreator := func() *app.App {
				cms := rootmulti.NewStore(appDB, ctx.Logger)
				cms.SetLazyLoading(true)
				return app.New(
					ctx.Logger, appDB, nil, false, nil,
					ctx.Config.RootDir, 0, encCfg, ctx.Viper,
					func(baseApp *baseapp.BaseApp) { baseApp.SetCMS(cms) },
				)
			}
			processBlock := func(height int64) (err error) {
				results, err := tmDB.stateStore.LoadABCIResponses(height)
				// fmt.Printf("mm-blockResult: %+v\n", results)
				if err != nil {
					return err
				}
				// rpcresult := &ctypes.ResultBlockResults{
				// 	Height:                int64(height),
				// 	TxsResults:            results.DeliverTxs,
				// 	BeginBlockEvents:      results.BeginBlock.Events,
				// 	EndBlockEvents:        results.EndBlock.Events,
				// 	ValidatorUpdates:      results.EndBlock.ValidatorUpdates,
				// 	ConsensusParamUpdates: results.EndBlock.ConsensusParamUpdates,
				// }
				// bz, err := json.Marshal(rpcresult)
				// if err != nil {
				// 	panic(err)
				// }
				// fmt.Println("mm-res:", string(bz))
				block := tmDB.blockStore.LoadBlock(height)
				// fmt.Printf("mm-block: %+v\n", block)
				if block == nil {
					return fmt.Errorf("block %d not found", height)
				}
				for txIndex, txResult := range results.DeliverTxs {
					tx := block.Txs[txIndex]
					// txHash := tx.Hash()
					app := appCreator()
					// existedRes := &abci.TxResult{
					// 	Height: block.Height,
					// 	Index:  uint32(txIndex),
					// 	Tx:     block.Txs[txIndex],
					// 	Result: *txResult,
					// }
					result, err := tmDB.replayTx(app, block, txIndex, state.InitialHeight)
					if err != nil {
						return err
					}
					// fmt.Printf("mm-existedRes: %+v\n", existedRes.Result)
					// fmt.Printf("mm-result: %+v\n", result.Result)
					// return nil
					parsedTx, err := clientCtx.TxConfig.TxDecoder()(result.Tx)
					if err != nil {
						fmt.Println("can't parse the patched tx", result.Height, result.Index)
						return err
					}
					for msgIndex, msg := range parsedTx.GetMsgs() {
						ethMsg, ok := msg.(*types.MsgEthereumTx)
						if !ok {
							continue
						}
						txHash := common.HexToHash(ethMsg.Hash)
						// Get transaction by hash
						txs, err := rpctypes.ParseTxResult(txResult, parsedTx)
						if err != nil {
							return fmt.Errorf("failed to parse tx events: block %d, index %d, %v", result.Height, result.Index, err)
						}
						newParsedTx := txs.GetTxByHash(txHash)
						if newParsedTx == nil {
							return fmt.Errorf("ethereum tx not found in msgs: block %d, index %d", result.Height, result.Index)
						}
						transaction := ethermint.TxResult{
							Height:            result.Height,
							TxIndex:           result.Index,
							MsgIndex:          uint32(msgIndex),
							EthTxIndex:        newParsedTx.EthTxIndex,
							Failed:            newParsedTx.Failed,
							GasUsed:           newParsedTx.GasUsed,
							CumulativeGasUsed: txs.AccumulativeGasUsed(newParsedTx.MsgIndex),
						}
						// check if block number is 0
						if transaction.Height == 0 {
							return errors.New("genesis is not traceable")
						}
						// check tx index is not out of bound
						if uint32(len(block.Txs)) < transaction.TxIndex {
							fmt.Println("tx index out of bounds", "index", transaction.TxIndex, "hash", txHash, "height", block.Height)
							return fmt.Errorf("transaction not included in block %v", block.Height)
						}

						var predecessors []*types.MsgEthereumTx
						for _, txBz := range block.Txs[:transaction.TxIndex] {
							tx, err := clientCtx.TxConfig.TxDecoder()(txBz)
							if err != nil {
								fmt.Println("failed to decode transaction in block", "height", block.Height, "error", err.Error())
								continue
							}
							for _, msg := range tx.GetMsgs() {
								ethMsg, ok := msg.(*types.MsgEthereumTx)
								if !ok {
									continue
								}

								predecessors = append(predecessors, ethMsg)
							}
						}

						blockTx, err := clientCtx.TxConfig.TxDecoder()(block.Txs[transaction.TxIndex])
						if err != nil {
							fmt.Println("mm-tx not found hash", txHash)
							return err
						}

						// add predecessor messages in current cosmos tx
						for i := 0; i < int(transaction.MsgIndex); i++ {
							ethMsg, ok := blockTx.GetMsgs()[i].(*types.MsgEthereumTx)
							if !ok {
								continue
							}
							predecessors = append(predecessors, ethMsg)
						}

						ethMessage, ok := blockTx.GetMsgs()[transaction.MsgIndex].(*types.MsgEthereumTx)

						if !ok {
							fmt.Println("invalid transaction type", "type", fmt.Sprintf("%T", tx))
							return fmt.Errorf("invalid transaction type %T", tx)
						}

						traceTxRequest := types.QueryTraceTxRequest{
							Msg:          ethMessage,
							Predecessors: predecessors,
							BlockNumber:  block.Height,
							BlockTime:    block.Time,
							BlockHash:    common.Bytes2Hex(block.Hash()),
						}
						// fmt.Printf("mm-traceTxRequest: %+v\n", traceTxRequest)
						contextHeight := transaction.Height - 1
						if contextHeight < 1 {
							contextHeight = 1
						}

						// 0xfef22c2b121d1e0ddd0a5e072ec8fea91d8c0ff2a6e37fb22443c1d74a324018
						proposerAddress := "A3F7397F81D4CC82D566633F9728DFB50A35D965"
						// 0xf699cf6cfc8b7c05ec123cdcaa2f1595f9c5bd757971a3ece189158722c49938
						if block.Height == 5697077 {
							proposerAddress = "FBCD779E246D1E43F935D960D04A5026823DD544"
						}
						proposerAdr, err := hex.DecodeString(proposerAddress)
						if err != nil {
							panic(err)
						}
						sdkCtx := app.NewContext(true, tmproto.Header{
							Height:          contextHeight,
							ProposerAddress: proposerAdr,
						})
						cfg, err := app.EvmKeeper.EVMConfig(sdkCtx, proposerAdr)
						if err != nil {
							fmt.Printf("mm-cfg-err: %+v\n", err)
							return err
						}
						txConfig := statedb.NewEmptyTxConfig(common.BytesToHash(sdkCtx.HeaderHash().Bytes()))
						signer := ethtypes.MakeSigner(cfg.ChainConfig, big.NewInt(sdkCtx.BlockHeight()))
						tx := traceTxRequest.Msg.AsTransaction()
						result, _, err := traceTx(sdkCtx, cfg, txConfig, signer, tx, traceTxRequest.TraceConfig, false, app)
						if err != nil {
							fmt.Printf("mm-traceTx-err: %+v\n", err)
							return err
						}

						resultData, err := json.Marshal(result)
						if err != nil {
							return status.Error(codes.Internal, err.Error())
						}
						return nil
						fmt.Printf("mm-decodedResult: %+v\n", string(resultData))
					}
				}
				return nil
			}
			concurrency, err := cmd.Flags().GetInt(FlagConcurrency)
			if err != nil {
				return err
			}

			blockChan := make(chan int64, concurrency)
			var wg sync.WaitGroup
			ctCtx, cancel := context.WithCancel(context.Background())
			defer cancel()

			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					defer wg.Done()
					for {
						select {
						case <-ctCtx.Done():
							return
						case blockNum, ok := <-blockChan:
							if !ok {
								return
							}

							if err := processBlock(blockNum); err != nil {
								fmt.Printf("error when processBlock: %d %+v\n", blockNum, err)
								cancel()
								return
							}
						}
					}
				}(&wg)
			}

			findBlock := func() error {
				startHeight, err := cmd.Flags().GetInt(FlagStartBlock)
				if err != nil {
					return err
				}
				endHeight, err := cmd.Flags().GetInt(FlagEndBlock)
				if err != nil {
					return err
				}
				if startHeight < 1 {
					return fmt.Errorf("invalid start-block: %d", startHeight)
				}
				if endHeight < startHeight {
					return fmt.Errorf("invalid end-block %d, smaller than start-block", endHeight)
				}

				for height := startHeight; height <= endHeight; height++ {
					blockChan <- int64(height)
				}
				return nil
			}

			go func() {
				err := findBlock()
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				close(blockChan)
			}()

			wg.Wait()

			return ctCtx.Err()
		},
	}
	cmd.Flags().String(flags.FlagChainID, "cronosmainnet_25-1", "network chain ID, only useful for psql tx indexer backend")
	cmd.Flags().Bool(flags.FlagDryRun, false, "Print the execution result of the problematic txs without patch the database")
	cmd.Flags().Int(FlagStartBlock, 1, "The start of the block range to iterate, inclusive")
	cmd.Flags().Int(FlagEndBlock, -1, "The end of the block range to iterate, inclusive")
	cmd.Flags().Int(FlagConcurrency, runtime.NumCPU(), "Define how many workers run in concurrency")
	return cmd
}

func openAppDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", dbm.RocksDBBackend, dataDir)
}

func openIndexerDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("evmindexer", dbm.RocksDBBackend, dataDir)
}

type tmDB struct {
	blockStore *tmstore.BlockStore
	stateStore sm.Store
	txIndexer  txindex.TxIndexer
}

func openTMDB(cfg *tmcfg.Config, chainID string) (*tmDB, error) {
	// open tendermint db
	tmdb, err := tmnode.DefaultDBProvider(&tmnode.DBContext{ID: "blockstore", Config: cfg})
	if err != nil {
		return nil, err
	}
	blockStore := tmstore.NewBlockStore(tmdb)
	stateDB, err := tmnode.DefaultDBProvider(&tmnode.DBContext{ID: "state", Config: cfg})
	if err != nil {
		return nil, err
	}
	stateStore := sm.NewStore(stateDB, sm.StoreOptions{
		DiscardABCIResponses: cfg.Storage.DiscardABCIResponses,
	})
	txIndexer, err := newTxIndexer(cfg, chainID)
	if err != nil {
		return nil, err
	}
	return &tmDB{
		blockStore, stateStore, txIndexer,
	}, nil
}

func newTxIndexer(config *tmcfg.Config, chainID string) (txindex.TxIndexer, error) {
	switch config.TxIndex.Indexer {
	case "kv":
		store, err := tmnode.DefaultDBProvider(&tmnode.DBContext{ID: "tx_index", Config: config})
		if err != nil {
			return nil, err
		}
		return kv.NewTxIndex(store), nil
	case "psql":
		if config.TxIndex.PsqlConn == "" {
			return nil, errors.New(`no psql-conn is set for the "psql" indexer`)
		}
		es, err := psql.NewEventSink(config.TxIndex.PsqlConn, chainID)
		if err != nil {
			return nil, fmt.Errorf("creating psql indexer: %w", err)
		}
		return es.TxIndexer(), nil
	default:
		return &null.TxIndex{}, nil
	}
}

func getBeginBlockValidatorInfo(block *tmtypes.Block, store sm.Store,
	initialHeight int64) abci.LastCommitInfo {
	voteInfos := make([]abci.VoteInfo, block.LastCommit.Size())
	// Initial block -> LastCommitInfo.Votes are empty.
	// Remember that the first LastCommit is intentionally empty, so it makes
	// sense for LastCommitInfo.Votes to also be empty.
	if block.Height > initialHeight {
		lastValSet, err := store.LoadValidators(block.Height - 1)
		if err != nil {
			panic(err)
		}

		// Sanity check that commit size matches validator set size - only applies
		// after first block.
		var (
			commitSize = block.LastCommit.Size()
			valSetLen  = len(lastValSet.Validators)
		)
		if commitSize != valSetLen {
			panic(fmt.Sprintf(
				"commit size (%d) doesn't match valset length (%d) at height %d\n\n%v\n\n%v",
				commitSize, valSetLen, block.Height, block.LastCommit.Signatures, lastValSet.Validators,
			))
		}

		for i, val := range lastValSet.Validators {
			commitSig := block.LastCommit.Signatures[i]
			voteInfos[i] = abci.VoteInfo{
				Validator:       tmtypes.TM2PB.Validator(val),
				SignedLastBlock: !commitSig.Absent(),
			}
		}
	}

	return abci.LastCommitInfo{
		Round: block.LastCommit.Round,
		Votes: voteInfos,
	}
}

// replay the tx and return the result
func (db *tmDB) replayTx(anApp *app.App, block *tmtypes.Block, txIndex int, initialHeight int64) (*abci.TxResult, error) {
	if err := anApp.LoadHeight(block.Height - 1); err != nil {
		return nil, err
	}

	pbh := block.Header.ToProto()
	if pbh == nil {
		return nil, errors.New("nil header")
	}

	byzVals := make([]abci.Evidence, 0)
	for _, evidence := range block.Evidence.Evidence {
		byzVals = append(byzVals, evidence.ABCI()...)
	}

	commitInfo := getBeginBlockValidatorInfo(block, db.stateStore, initialHeight)

	_ = anApp.BeginBlock(abci.RequestBeginBlock{
		Hash:                block.Hash(),
		Header:              *pbh,
		ByzantineValidators: byzVals,
		LastCommitInfo:      commitInfo,
	})

	// run the predecessor txs
	for _, tx := range block.Txs[:txIndex] {
		anApp.DeliverTx(abci.RequestDeliverTx{Tx: tx})
	}
	rsp := anApp.DeliverTx(abci.RequestDeliverTx{Tx: block.Txs[txIndex]})
	return &abci.TxResult{
		Height: block.Height,
		Index:  uint32(txIndex),
		Tx:     block.Txs[txIndex],
		Result: rsp,
	}, nil
}

func traceTx(
	ctx sdk.Context,
	cfg *types.EVMConfig,
	txConfig statedb.TxConfig,
	signer ethtypes.Signer,
	tx *ethtypes.Transaction,
	traceConfig *types.TraceConfig,
	commitMessage bool,
	app *app.App,
) (*interface{}, uint, error) {

	// Assemble the structured logger or the JavaScript tracer
	var (
		tracer    tracers.Tracer
		overrides *params.ChainConfig
		err       error
	)
	msg, err := tx.AsMessage(signer, cfg.BaseFee)
	if err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	if traceConfig == nil {
		traceConfig = &types.TraceConfig{}
	}

	if traceConfig.Overrides != nil {
		overrides = traceConfig.Overrides.EthereumConfig(cfg.ChainConfig.ChainID)
	}

	logConfig := logger.Config{
		EnableMemory:     traceConfig.EnableMemory,
		DisableStorage:   traceConfig.DisableStorage,
		DisableStack:     traceConfig.DisableStack,
		EnableReturnData: traceConfig.EnableReturnData,
		Debug:            traceConfig.Debug,
		Limit:            int(traceConfig.Limit),
		Overrides:        overrides,
	}

	tracer = logger.NewStructLogger(&logConfig)

	tCtx := &tracers.Context{
		BlockHash: txConfig.BlockHash,
		TxIndex:   int(txConfig.TxIndex),
		TxHash:    txConfig.TxHash,
	}

	if traceConfig.Tracer != "" {
		if tracer, err = tracers.New(traceConfig.Tracer, tCtx); err != nil {
			return nil, 0, status.Error(codes.Internal, err.Error())
		}
	}

	res, err := applyMessageWithConfig(app, ctx, msg, tracer, commitMessage, cfg, txConfig)
	if err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	var result interface{}
	result, err = tracer.GetResult()
	if err != nil {
		return nil, 0, status.Error(codes.Internal, err.Error())
	}

	return &result, txConfig.LogIndex + uint(len(res.Logs)), nil
}

func getEthIntrinsicGas(ctx sdk.Context, msg core.Message, cfg *params.ChainConfig, isContractCreation bool) (uint64, error) {
	height := big.NewInt(ctx.BlockHeight())
	homestead := cfg.IsHomestead(height)
	istanbul := cfg.IsIstanbul(height)

	return core.IntrinsicGas(msg.Data(), msg.AccessList(), isContractCreation, homestead, istanbul)
}

func gasToRefund(availableRefund, gasConsumed, refundQuotient uint64) uint64 {
	// Apply refund counter
	refund := gasConsumed / refundQuotient
	if refund > availableRefund {
		return availableRefund
	}
	return refund
}

func GetMinGasMultiplier(app *app.App, ctx sdk.Context) sdk.Dec {
	fmkParmas := app.FeeMarketKeeper.GetParams(ctx)
	if fmkParmas.MinGasMultiplier.IsNil() {
		// in case we are executing eth_call on a legacy block, returns a zero value.
		return sdk.ZeroDec()
	}
	return fmkParmas.MinGasMultiplier
}

func applyMessageWithConfig(
	app *app.App,
	ctx sdk.Context,
	msg core.Message,
	tracer vm.EVMLogger,
	commit bool,
	cfg *types.EVMConfig,
	txConfig statedb.TxConfig,
) (*types.MsgEthereumTxResponse, error) {
	var (
		ret   []byte // return bytes from evm execution
		vmErr error  // vm errors do not effect consensus and are therefore not assigned to err
	)

	// return error if contract creation or call are disabled through governance
	if !cfg.Params.EnableCreate && msg.To() == nil {
		return nil, sdkerrors.Wrap(types.ErrCreateDisabled, "failed to create new contract")
	} else if !cfg.Params.EnableCall && msg.To() != nil {
		return nil, sdkerrors.Wrap(types.ErrCallDisabled, "failed to call contract")
	}

	stateDB := statedb.New(ctx, app.EvmKeeper, txConfig)
	evm := app.EvmKeeper.NewEVM(ctx, msg, cfg, tracer, stateDB)

	leftoverGas := msg.Gas()

	// Allow the tracer captures the tx level events, mainly the gas consumption.
	vmCfg := evm.Config()
	if vmCfg.Debug {
		vmCfg.Tracer.CaptureTxStart(leftoverGas)
		defer func() {
			vmCfg.Tracer.CaptureTxEnd(leftoverGas)
		}()
	}

	sender := vm.AccountRef(msg.From())
	contractCreation := msg.To() == nil
	isLondon := cfg.ChainConfig.IsLondon(evm.Context().BlockNumber)

	intrinsicGas, err := getEthIntrinsicGas(ctx, msg, cfg.ChainConfig, contractCreation)
	if err != nil {
		// should have already been checked on Ante Handler
		return nil, sdkerrors.Wrap(err, "intrinsic gas failed")
	}

	// Should check again even if it is checked on Ante Handler, because eth_call don't go through Ante Handler.
	if leftoverGas < intrinsicGas {
		// eth_estimateGas will check for this exact error
		return nil, sdkerrors.Wrap(core.ErrIntrinsicGas, "apply message")
	}
	leftoverGas -= intrinsicGas

	// access list preparation is moved from ante handler to here, because it's needed when `ApplyMessage` is called
	// under contexts where ante handlers are not run, for example `eth_call` and `eth_estimateGas`.
	if rules := cfg.ChainConfig.Rules(big.NewInt(ctx.BlockHeight()), cfg.ChainConfig.MergeNetsplitBlock != nil); rules.IsBerlin {
		stateDB.PrepareAccessList(msg.From(), msg.To(), evm.ActivePrecompiles(rules), msg.AccessList())
	}

	if contractCreation {
		// take over the nonce management from evm:
		// - reset sender's nonce to msg.Nonce() before calling evm.
		// - increase sender's nonce by one no matter the result.
		stateDB.SetNonce(sender.Address(), msg.Nonce())
		ret, _, leftoverGas, vmErr = evm.Create(sender, msg.Data(), leftoverGas, msg.Value())
		stateDB.SetNonce(sender.Address(), msg.Nonce()+1)
	} else {
		ret, leftoverGas, vmErr = evm.Call(sender, *msg.To(), msg.Data(), leftoverGas, msg.Value())
	}

	refundQuotient := params.RefundQuotient

	// After EIP-3529: refunds are capped to gasUsed / 5
	if isLondon {
		refundQuotient = params.RefundQuotientEIP3529
	}

	// calculate gas refund
	if msg.Gas() < leftoverGas {
		return nil, sdkerrors.Wrap(types.ErrGasOverflow, "apply message")
	}
	// refund gas
	re := stateDB.GetRefund()
	leftoverGas += gasToRefund(re, msg.Gas()-leftoverGas, refundQuotient)
	// EVM execution error needs to be available for the JSON-RPC client
	var vmError string
	if vmErr != nil {
		vmError = vmErr.Error()
	}

	// The dirty states in `StateDB` is either committed or discarded after return
	if commit {
		if err := stateDB.Commit(); err != nil {
			return nil, sdkerrors.Wrap(err, "failed to commit stateDB")
		}
	}

	// calculate a minimum amount of gas to be charged to sender if GasLimit
	// is considerably higher than GasUsed to stay more aligned with Tendermint gas mechanics
	// for more info https://github.com/evmos/ethermint/issues/1085
	gasLimit := sdk.NewDec(int64(msg.Gas()))
	minGasMultiplier := GetMinGasMultiplier(app, ctx)
	minimumGasUsed := gasLimit.Mul(minGasMultiplier)
	if msg.Gas() < leftoverGas {
		return nil, sdkerrors.Wrapf(types.ErrGasOverflow, "message gas limit < leftover gas (%d < %d)", msg.Gas(), leftoverGas)
	}
	temporaryGasUsed := msg.Gas() - leftoverGas
	gasUsed := sdk.MaxDec(minimumGasUsed, sdk.NewDec(int64(temporaryGasUsed))).TruncateInt().Uint64()
	// reset leftoverGas, to be used by the tracer
	leftoverGas = msg.Gas() - gasUsed
	return &types.MsgEthereumTxResponse{
		GasUsed: gasUsed,
		VmError: vmError,
		Ret:     ret,
		Logs:    types.NewLogsFromEth(stateDB.Logs()),
		Hash:    txConfig.TxHash.Hex(),
	}, nil
}
