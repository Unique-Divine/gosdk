package gonibi

import (
	"context"
	"errors"

	xwasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/NibiruChain/nibiru/app"
	xepochs "github.com/NibiruChain/nibiru/x/epochs/types"
	xoracle "github.com/NibiruChain/nibiru/x/oracle/types"
	xperp "github.com/NibiruChain/nibiru/x/perp/v2/types"
	cmtrpcclient "github.com/cometbft/cometbft/rpc/client"
	"google.golang.org/grpc"

	"github.com/Unique-Divine/gonibi/cmdctx"
	sdkclient "github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	csdk "github.com/cosmos/cosmos-sdk/types"
	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type INibiruClient interface {
	BroadcastMsgs(from csdk.AccAddress, msgs ...csdk.Msg) (*csdk.TxResponse, error)
	ClientCtx(tmCfgRootDir string) sdkclient.Context
	GetAccountNumbers(address string) (nums AccountNumbers, err error)
}

var _ INibiruClient = (*NibiruClient)(nil)

type NibiruClient struct {
	ChainId          string
	Keyring          keyring.Keyring
	EncCfg           app.EncodingConfig
	Tx               GrpcBroadcastClient
	Querier          Querier
	CometRPC         cmtrpcclient.Client
	AccountRetriever authtypes.AccountRetriever
	GrpcClient       *grpc.ClientConn
}

func NewNibiruClient(
	chainId string,
	grpcConn *grpc.ClientConn,
	rpcEndpt string,
) (NibiruClient, error) {
	EnsureNibiruPrefix()
	encCfg := app.MakeEncodingConfig()
	keyring := keyring.NewInMemory(encCfg.Marshaler)
	queryClient, err := NewQueryClient(grpcConn)
	if err != nil {
		return NibiruClient{}, err
	}
	cometRpc, err := NewRPCClient(rpcEndpt, "/websocket")
	if err != nil {
		return NibiruClient{}, err
	}
	accRetriever := authtypes.AccountRetriever{}
	return NibiruClient{
		ChainId: chainId,
		Keyring: keyring,
		EncCfg:  encCfg,
		Tx: GrpcBroadcastClient{
			ServiceClient: cosmostx.NewServiceClient(grpcConn),
		},
		Querier:          queryClient,
		CometRPC:         cometRpc,
		AccountRetriever: accRetriever,
		GrpcClient:       grpcConn,
	}, err
}

// ClientCtx: Docs for args TODO
//
//   - tmCfgRootDir: /node0/simd
//   - Validator.Dir: /node0
//   - Validator.ClientCtx.KeyringDir: /node0/simcli
func (nc *NibiruClient) ClientCtx(
	tmCfgRootDir string,
) sdkclient.Context {
	encCfg := nc.EncCfg
	return cmdctx.NewClientCtx(cmdctx.ClientCtxBuilder{
		Keyring:           nc.Keyring,
		TmCfgRootDir:      tmCfgRootDir, // Not sure what to put here
		ChainID:           nc.ChainId,
		AccountRetriever:  nc.AccountRetriever,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
	})
}

func EnsureNibiruPrefix() {
	csdkConfig := csdk.GetConfig()
	nibiruPrefix := app.AccountAddressPrefix
	if csdkConfig.GetBech32AccountAddrPrefix() != nibiruPrefix {
		app.SetPrefixes(nibiruPrefix)
	}
}

type Querier struct {
	ClientConn *grpc.ClientConn
	Perp       xperp.QueryClient
	Epoch      xepochs.QueryClient
	Oracle     xoracle.QueryClient
	Wasm       xwasm.QueryClient
}

func NewQueryClient(
	grpcConn *grpc.ClientConn,
) (Querier, error) {
	if grpcConn == nil {
		return Querier{}, errors.New(
			"cannot create NibiruQueryClient with nil grpc.ClientConn")
	}

	return Querier{
		ClientConn: grpcConn,
		Perp:       xperp.NewQueryClient(grpcConn),
		Epoch:      xepochs.NewQueryClient(grpcConn),
		Oracle:     xoracle.NewQueryClient(grpcConn),
		Wasm:       xwasm.NewQueryClient(grpcConn),
	}, nil
}

type AccountNumbers struct {
	Number   uint64
	Sequence uint64
}

func GetAccountNumbers(
	address string,
	grpcConn *grpc.ClientConn,
	encCfg app.EncodingConfig,
) (nums AccountNumbers, err error) {
	queryClient := authtypes.NewQueryClient(grpcConn)
	resp, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return nums, err
	}
	// register auth interface

	var acc authtypes.AccountI
	encCfg.InterfaceRegistry.UnpackAny(resp.Account, &acc)

	return AccountNumbers{
		Number:   acc.GetAccountNumber(),
		Sequence: acc.GetSequence(),
	}, err
}

func (nc *NibiruClient) GetAccountNumbers(
	address string,
) (nums AccountNumbers, err error) {
	return GetAccountNumbers(address, nc.Querier.ClientConn, nc.EncCfg)
}

type GrpcBroadcastClient struct {
	cosmostx.ServiceClient
}
