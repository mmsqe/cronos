package cmd

import (
	"bytes"
	"log"
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/crypto-org-chain/cronos/app"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/protoio"
	tmstate "github.com/tendermint/tendermint/proto/tendermint/state"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/state/txindex/kv"
	tmstore "github.com/tendermint/tendermint/store"
	"github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
)

func mockBlockResult() *tmstate.ABCIResponses {
	return &tmstate.ABCIResponses{
		DeliverTxs: []*abci.ResponseDeliverTx{
			{Code: 0, Data: []byte{0x01}, Log: "ok"},
			{Code: 1, Log: "not ok"},
		},
		EndBlock:   &abci.ResponseEndBlock{},
		BeginBlock: &abci.ResponseBeginBlock{},
	}
}

func mockResult(txGen client.TxConfig) *abci.TxResult {
	txs, _ := app.GenSequenceOfTxs(txGen, nil, nil, nil, 1)
	txBytes, _ := txGen.TxEncoder()(txs[0])
	return &abci.TxResult{
		Height: 1,
		Index:  0,
		Tx:     txBytes,
		Result: abci.ResponseDeliverTx{
			Code: abci.CodeTypeOK,
		},
	}
}

func getExpected(result *abci.TxResult, blockRes *tmstate.ABCIResponses) []byte {
	expected := new(bytes.Buffer)
	protoWriter := protoio.NewDelimitedWriter(expected)
	results := make([]proto.Message, 0)
	if result != nil {
		results = append(results, result)
	}
	if blockRes != nil {
		results = append(results, blockRes)
	}
	for _, res := range results {
		_, err := protoWriter.WriteMsg(res)
		if err != nil {
			log.Fatal(err)
		}
	}
	protoWriter.Close()
	return expected.Bytes()
}

func mockTmDb() *tmDB {
	db := tmdb.NewMemDB()
	return &tmDB{
		blockStore: tmstore.NewBlockStore(db),
		stateStore: sm.NewStore(db),
		txIndexer:  kv.NewTxIndex(db),
	}
}

func TestPatchToExport(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	tmDB := mockTmDb()
	t.Run("TestPatchToExport", func(t *testing.T) {
		blockRes := mockBlockResult()
		res := mockResult(encCfg.TxConfig)
		expected := getExpected(res, blockRes)
		b := new(bytes.Buffer)
		err := tmDB.PatchToExport(blockRes, res, b)
		require.NoError(t, err)
		require.Equal(t, b.Bytes(), expected)
	})
}

func TestPatchFromImport(t *testing.T) {
	tmDB := mockTmDb()
	encCfg := simapp.MakeTestEncodingConfig()

	t.Run("happy flow", func(t *testing.T) {
		res := mockResult(encCfg.TxConfig)
		blockRes := mockBlockResult()
		blockRes.DeliverTxs[res.Index] = &res.Result
		expected := getExpected(res, blockRes)
		err := tmDB.PatchFromImport(encCfg.TxConfig, bytes.NewReader(expected))
		require.NoError(t, err, "import error")
		txHash := types.Tx(res.Tx).Hash()
		newRes, err := tmDB.txIndexer.Get(txHash)
		require.NoError(t, err, "get tx result")
		resultProto, _ := res.Marshal()
		newResProto, _ := newRes.Marshal()
		require.Equal(t, resultProto, newResProto, "check tx result")
		newBlockRes, err := tmDB.stateStore.LoadABCIResponses(res.Height)
		require.NoError(t, err, "get block rseult")
		blockResProto, _ := blockRes.Marshal()
		newBlockResProto, _ := newBlockRes.Marshal()
		require.Equal(t, blockResProto, newBlockResProto, "check block result")
	})

	t.Run("wrong object type", func(t *testing.T) {
		blockRes := mockBlockResult()
		expected := getExpected(nil, blockRes)
		err := tmDB.PatchFromImport(encCfg.TxConfig, bytes.NewReader(expected))
		require.EqualError(t, err, "proto: wrong wireType = 2 for field Height")
	})

	t.Run("wrong last object", func(t *testing.T) {
		res := mockResult(encCfg.TxConfig)
		expected := getExpected(res, nil)
		err := tmDB.PatchFromImport(encCfg.TxConfig, bytes.NewReader(expected))
		require.EqualError(t, err, "EOF")
	})
}
