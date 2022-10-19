package v2_test

import (
	"testing"

	"github.com/crypto-org-chain/cronos/x/cronos/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/crypto-org-chain/cronos/app"
)

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

type MigrationTestSuite struct {
	suite.Suite

	ctx sdk.Context
	app *app.App

	params types.Params
}

func (suite *MigrationTestSuite) SetupTest() {
	// account key
	priv, err := ethsecp256k1.GenerateKey()
	require.NoError(suite.T(), err)
	address := common.BytesToAddress(priv.PubKey().Address().Bytes())
	cronosAdmin := sdk.AccAddress(address.Bytes()).String()

	suite.app = app.Setup(suite.T(), cronosAdmin, true)
	suite.ctx = suite.app.NewContext(false, tmproto.Header{})
	suite.params = types.DefaultParams()
	suite.params.CronosAdmin = cronosAdmin
	suite.params.EnableAutoDeployment = true
}

func (suite *MigrationTestSuite) TestMigrate() {
	suite.Require().Equal(suite.params, suite.app.CronosKeeper.GetParams(suite.ctx))
}
