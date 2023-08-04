package gonibi_test

import (
	"fmt"
	"testing"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil"
	"github.com/NibiruChain/nibiru/x/common/testutil/cli"
	"github.com/NibiruChain/nibiru/x/common/testutil/genesis"
	"github.com/Unique-Divine/gonibi"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	tmconfig "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// --------------------------------------------------
// NibiruClientSuite
// --------------------------------------------------

type NibiruClientSuite struct {
	suite.Suite

	gosdk    *gonibi.NibiruClient
	grpcConn *grpc.ClientConn
	cfg      *cli.Config
	network  *cli.Network
	val      *cli.Validator
}

func TestNibiruClientTestSuite_RunAll(t *testing.T) {
	suite.Run(t, new(NibiruClientSuite))
}

func AbsorbServerConfig(
	cfg *cli.Config, srvCfg *serverconfig.Config,
) *cli.Config {
	cfg.GRPCAddress = srvCfg.GRPC.Address
	cfg.APIAddress = srvCfg.API.Address
	return cfg
}

func AbsorbTmConfig(
	cfg *cli.Config, tmCfg *tmconfig.Config,
) *cli.Config {
	cfg.RPCAddress = tmCfg.RPC.ListenAddress
	return cfg
}

type Blockchain struct {
	GrpcConn *grpc.ClientConn
	Cfg      *cli.Config
	Network  *cli.Network
	Val      *cli.Validator
}

func CreateBlockchain(t *testing.T) (nibiru Blockchain, err error) {
	gonibi.EnsureNibiruPrefix()
	encConfig := app.MakeEncodingConfig()
	genState := genesis.NewTestGenesisState(encConfig)
	cliCfg := cli.BuildNetworkConfig(genState)
	cfg := &cliCfg
	cfg.NumValidators = 1

	network, err := cli.New(
		t,
		t.TempDir(),
		*cfg,
	)
	if err != nil {
		return nibiru, err
	}
	err = network.WaitForNextBlock()
	if err != nil {
		return nibiru, err
	}

	val := network.Validators[0]
	AbsorbServerConfig(cfg, val.AppConfig)
	AbsorbTmConfig(cfg, val.Ctx.Config)

	grpcConn, err := ConnectGrpcToVal(val)
	if err != nil {
		return nibiru, err
	}
	return Blockchain{
		GrpcConn: grpcConn,
		Cfg:      cfg,
		Network:  network,
		Val:      val,
	}, err
}

func (s *NibiruClientSuite) RPCEndpoint() string {
	return s.val.RPCAddress
}

func (s *NibiruClientSuite) SetupSuite() {
	nibiru, err := CreateBlockchain(s.T())
	s.NoError(err)
	s.network = nibiru.Network
	s.cfg = nibiru.Cfg
	s.val = nibiru.Val
	s.grpcConn = nibiru.GrpcConn
}

func ConnectGrpcToVal(val *cli.Validator) (*grpc.ClientConn, error) {
	grpcUrl := val.AppConfig.GRPC.Address
	return gonibi.GetGRPCConnection(
		grpcUrl, true, 5,
	)
}

func (s *NibiruClientSuite) ConnectGrpc() {
	grpcConn, err := ConnectGrpcToVal(s.val)
	s.NoError(err)
	s.NotNil(grpcConn)
	s.grpcConn = grpcConn
}

func (s *NibiruClientSuite) TestNewQueryClient() {
	_, err := gonibi.NewQueryClient(s.grpcConn)
	s.NoError(err)
}

func (s *NibiruClientSuite) TestNewNibiruClient() {

	gosdk, err := gonibi.NewNibiruClient(s.cfg.ChainID, s.grpcConn, s.RPCEndpoint())
	s.NoError(err)
	s.gosdk = &gosdk

	s.gosdk.Keyring = s.val.ClientCtx.Keyring
	s.T().Run("DoTestBroadcastMsgs", func(t *testing.T) {
		s.DoTestBroadcastMsgs()
	})
}

func (s *NibiruClientSuite) UsefulPrints() {
	tmCfgRootDir := s.val.Ctx.Config.RootDir
	fmt.Printf("tmCfgRootDir: %v\n", tmCfgRootDir)
	fmt.Printf("s.val.Dir: %v\n", s.val.Dir)
	fmt.Printf("s.val.ClientCtx.KeyringDir: %v\n", s.val.ClientCtx.KeyringDir)
}

func (s *NibiruClientSuite) DoTestBroadcastMsgs() {
	from := s.val.Address
	to := testutil.AccAddress()
	amt := sdk.NewCoins(sdk.NewInt64Coin(denoms.NIBI, 420))
	txResp, err := s.gosdk.BroadcastMsgs(
		from,
		banktypes.NewMsgSend(from, to, amt),
	)
	s.NoError(err)
	s.NotNil(txResp)
	s.EqualValues(txResp.Code, 0)
}

func (s *NibiruClientSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// --------------------------------------------------
// NibiruClientSuite_NoNetwork
// --------------------------------------------------

type NibiruClientSuite_NoNetwork struct {
	suite.Suite
}

func TestNibiruClientSuite_NoNetwork_RunAll(t *testing.T) {
	suite.Run(t, new(NibiruClientSuite_NoNetwork))
}

func (s *NibiruClientSuite_NoNetwork) TestGetGrpcConnection_NoNetwork() {
	grpcConn, err := gonibi.GetGRPCConnection(
		gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2,
	)
	s.Error(err)
	s.Nil(grpcConn)

	_, err = gonibi.NewQueryClient(grpcConn)
	s.Error(err)
}
