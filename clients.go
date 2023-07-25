package gonibi

import (
	"errors"

	xwasm "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/NibiruChain/nibiru/app"
	xepochs "github.com/NibiruChain/nibiru/x/epochs/types"
	xoracle "github.com/NibiruChain/nibiru/x/oracle/types"
	xperp "github.com/NibiruChain/nibiru/x/perp/v2/types"
	"google.golang.org/grpc"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	csdk "github.com/cosmos/cosmos-sdk/types"
	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
)

type NibiruClient struct {
	ChainId string
	Keyring keyring.Keyring
	encCfg  app.EncodingConfig
	Tx      BroadcastClient
	Query   QueryClient
}

func NewNibiruClient(chainId string, grpcConn *grpc.ClientConn) (NibiruClient, error) {
	ensureNibiruPrefix()
	encCfg := app.MakeEncodingConfig()
	keyring := keyring.NewInMemory(encCfg.Marshaler)
	queryClient, err := NewQueryClient(grpcConn)
	if err != nil {
		return NibiruClient{}, err
	}
	return NibiruClient{
		ChainId: chainId,
		Keyring: keyring,
		encCfg:  encCfg,
		Tx: BroadcastClient{
			ServiceClient: cosmostx.NewServiceClient(grpcConn),
		},
		Query: queryClient,
	}, err
}

func ensureNibiruPrefix() {
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

type BroadcastClient struct {
	cosmostx.ServiceClient
}

func (txclient *BroadcastClient) BroadcastMsgs() {
	// TODO
}
