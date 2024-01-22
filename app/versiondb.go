//go:build rocksdb
// +build rocksdb

package app

import (
	"os"
	"path/filepath"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/crypto-org-chain/cronos/v2/cmd/cronosd/opendb"
	"github.com/crypto-org-chain/cronos/versiondb"
	versiondbclient "github.com/crypto-org-chain/cronos/versiondb/client"
	"github.com/crypto-org-chain/cronos/versiondb/tsrocksdb"
	"github.com/linxGnu/grocksdb"
)

func (app *App) setupVersionDB(
	homePath string,
	keys map[string]*storetypes.KVStoreKey,
	tkeys map[string]*storetypes.TransientStoreKey,
	memKeys map[string]*storetypes.MemoryStoreKey,
) (sdk.MultiStore, versiondb.VersionStore, error) {
	dataDir := filepath.Join(homePath, "data", "versiondb")
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, nil, err
	}
	versionDB, err := tsrocksdb.NewStore(dataDir)
	if err != nil {
		return nil, nil, err
	}

	// default to exposing all
	exposeStoreKeys := make([]storetypes.StoreKey, 0, len(keys))
	for _, storeKey := range keys {
		exposeStoreKeys = append(exposeStoreKeys, storeKey)
	}

	service := versiondb.NewStreamingService(versionDB, exposeStoreKeys)
	app.SetStreamingService(service)

	verDB := versiondb.NewMultiStore(app.CommitMultiStore(), versionDB, exposeStoreKeys)
	verDB.MountTransientStores(tkeys)
	verDB.MountMemoryStores(memKeys)

	app.SetQueryMultiStore(verDB)
	return verDB, versionDB, nil
}

func GetOptions(storeNames []string) versiondbclient.Options {
	return versiondbclient.Options{
		DefaultStores:  storeNames,
		OpenReadOnlyDB: opendb.OpenReadOnlyDB,
		AppRocksDBOptions: func(sstFileWriter bool) *grocksdb.Options {
			return opendb.NewRocksdbOptions(nil, sstFileWriter)
		},
	}
}
