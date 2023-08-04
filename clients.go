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

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	sdkcodectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	csdk "github.com/cosmos/cosmos-sdk/types"
	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type NibiruClient struct {
	ChainId          string
	Keyring          keyring.Keyring
	EncCfg           app.EncodingConfig
	Tx               GrpcBroadcastClient
	Query            QueryClient
	CometRPC         cmtrpcclient.Client
	AccountRetriever authtypes.AccountRetriever
	GrpcClient       *grpc.ClientConn
}

func NewNibiruClient(chainId string, grpcConn *grpc.ClientConn, rpcEndpt string) (NibiruClient, error) {
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
		Query:            queryClient,
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
	keyringDir string,
	tmCfgRootDir string,
) sdkclient.Context {
	encCfg := nc.EncCfg
	return NewClientCtx(ClientCtxBuilder{
		KeyringDir:        keyringDir, // Not sure what to put here
		Keyring:           nc.Keyring,
		TmCfgRootDir:      tmCfgRootDir, // Not sure what to put here
		ChainID:           nc.ChainId,
		AccountRetriever:  nc.AccountRetriever,
		InterfaceRegistry: encCfg.InterfaceRegistry,
		Codec:             encCfg.Marshaler,
		TxConfig:          encCfg.TxConfig,
	})
}

type ClientCtxBuilder struct {
	KeyringDir       string
	Keyring          keyring.Keyring
	TmCfgRootDir     string
	ChainID          string
	AccountRetriever sdkclient.AccountRetriever
	// app.EncodingConfig.InterfaceRegistry
	InterfaceRegistry sdkcodectypes.InterfaceRegistry
	// app.EncodingConfig.Codec
	Codec sdkcodec.Codec
	// app.EncodingConfig.TxConfig
	TxConfig sdkclient.TxConfig
}

func NewClientCtx(args ClientCtxBuilder) sdkclient.Context {
	return sdkclient.Context{}.
		WithKeyringDir(args.KeyringDir).
		WithKeyring(args.Keyring).
		WithHomeDir(args.TmCfgRootDir).
		WithChainID(args.ChainID).
		WithInterfaceRegistry(args.InterfaceRegistry).
		WithCodec(args.Codec).
		WithTxConfig(args.TxConfig).
		WithAccountRetriever(args.AccountRetriever)
}

func EnsureNibiruPrefix() {
	csdkConfig := csdk.GetConfig()
	nibiruPrefix := app.AccountAddressPrefix
	if csdkConfig.GetBech32AccountAddrPrefix() != nibiruPrefix {
		app.SetPrefixes(nibiruPrefix)
	}
}

type QueryClient struct {
	ClientConn *grpc.ClientConn
	Perp       xperp.QueryClient
	Epoch      xepochs.QueryClient
	Oracle     xoracle.QueryClient
	Wasm       xwasm.QueryClient
}

func NewQueryClient(
	grpcConn *grpc.ClientConn,
) (QueryClient, error) {
	if grpcConn == nil {
		return QueryClient{}, errors.New(
			"cannot create NibiruQueryClient with nil grpc.ClientConn")
	}

	return QueryClient{
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

func (nc *NibiruClient) GetAccountNumbers(address string) (nums AccountNumbers, err error) {
	queryClient := authtypes.NewQueryClient(nc.GrpcClient)
	resp, err := queryClient.Account(context.Background(), &authtypes.QueryAccountRequest{
		Address: address,
	})
	if err != nil {
		return nums, err
	}
	// register auth interface

	var acc authtypes.AccountI
	nc.EncCfg.InterfaceRegistry.UnpackAny(resp.Account, &acc)

	return AccountNumbers{
		Number:   acc.GetAccountNumber(),
		Sequence: acc.GetSequence(),
	}, err
}

type GrpcBroadcastClient struct {
	cosmostx.ServiceClient
}
