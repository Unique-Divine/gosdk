package gonibi_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/NibiruChain/nibiru/x/common/denoms"
	"github.com/NibiruChain/nibiru/x/common/testutil"
	"github.com/NibiruChain/nibiru/x/common/testutil/cli"
	"github.com/Unique-Divine/gonibi"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

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

func (s *NibiruClientSuite) RPCEndpoint() string {
	return s.val.RPCAddress
}

func (s *NibiruClientSuite) SetupSuite() {
	nibiru, err := gonibi.CreateBlockchain(s.T())
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
	rpcEndpt := s.val.RPCAddress
	gosdk, err := gonibi.NewNibiruClient(s.cfg.ChainID, s.grpcConn, rpcEndpt)
	s.NoError(err)
	s.gosdk = &gosdk

	s.gosdk.Keyring = s.val.ClientCtx.Keyring
	s.T().Run("DoTestBroadcastMsgs", func(t *testing.T) {
		s.DoTestBroadcastMsgs()
	})
	s.T().Run("DoTestBroadcastMsgsGrpc", func(t *testing.T) {
		s.DoTestBroadcastMsgsGrpc()
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

func (s *NibiruClientSuite) DoTestBroadcastMsgsGrpc() {
	s.NoError(s.network.WaitForNextBlock())
	from := s.val.Address
	to := testutil.AccAddress()
	amt := sdk.NewCoins(sdk.NewInt64Coin(denoms.NIBI, 420))
	txResp, err := s.gosdk.BroadcastMsgsGrpc(
		from,
		banktypes.NewMsgSend(from, to, amt),
	)
	s.NoError(err)
	s.NotNil(txResp)
	base := 10
	var txRespCode string = strconv.FormatUint(uint64(txResp.Code), base)
	s.EqualValuesf(txResp.Code, 0,
		"code: %v\nraw log: %s", txRespCode, txResp.RawLog)
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
