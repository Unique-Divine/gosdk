package gonibi_test

import (
	"github.com/Unique-Divine/gonibi"
	"testing"

	"github.com/NibiruChain/nibiru/app"
	"github.com/NibiruChain/nibiru/x/common/testutil/cli"
	"github.com/NibiruChain/nibiru/x/common/testutil/genesis"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	tmconfig "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// --------------------------------------------------
// GrpcClientSuite
// --------------------------------------------------

type GrpcClientSuite struct {
	suite.Suite

	gosdk    *gonibi.NibiruClient
	grpcConn *grpc.ClientConn
	cfg      *cli.Config
	network  *cli.Network
	val      *cli.Validator
}

func TestGrpcClientTestSuite_RunAll(t *testing.T) {
	suite.Run(t, new(GrpcClientSuite))
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

func (s *GrpcClientSuite) SetupSuite() {
	app.SetPrefixes(app.AccountAddressPrefix)
	encConfig := app.MakeEncodingConfig()
	genState := genesis.NewTestGenesisState(encConfig)
	cliCfg := cli.BuildNetworkConfig(genState)
	s.cfg = &cliCfg
	s.cfg.NumValidators = 1

	network, err := cli.New(
		s.T(),
		s.T().TempDir(),
		*s.cfg,
	)
	s.NoError(err)
	s.network = network
	s.NoError(s.network.WaitForNextBlock())

	s.val = s.network.Validators[0]
	AbsorbServerConfig(s.cfg, s.val.AppConfig)
	AbsorbTmConfig(s.cfg, s.val.Ctx.Config)
	s.ConnectGrpc()
}

func (s *GrpcClientSuite) ConnectGrpc() {
	grpcUrl := s.val.AppConfig.GRPC.Address
	grpcConn, err := gonibi.GetGRPCConnection(
		grpcUrl, true, 5,
	)
	s.NoError(err)
	s.NotNil(grpcConn)
	s.grpcConn = grpcConn
}

func (s *GrpcClientSuite) TestNewQueryClient() {
	_, err := gonibi.NewQueryClient(s.grpcConn)
	s.NoError(err)
}

func (s *GrpcClientSuite) TestNewNibiruClient() {
	_, err := gonibi.NewNibiruClient(s.cfg.ChainID, s.grpcConn)
	s.NoError(err)
}

func (s *GrpcClientSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// --------------------------------------------------
// GrpcClientSuite_NoNetwork
// --------------------------------------------------

type GrpcClientSuite_NoNetwork struct {
	suite.Suite
}

func TestGrpcClientSuite_NoNetwork_RunAll(t *testing.T) {
	suite.Run(t, new(GrpcClientSuite_NoNetwork))
}

func (s *GrpcClientSuite_NoNetwork) TestGetGrpcConnection_NoNetwork() {
	grpcConn, err := gonibi.GetGRPCConnection(
		gonibi.DefaultNetworkInfo.GrpcEndpoint, true, 2,
	)
	s.Error(err)
	s.Nil(grpcConn)

	_, err = gonibi.NewQueryClient(grpcConn)
	s.Error(err)
}
