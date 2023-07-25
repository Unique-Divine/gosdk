package gonibi

import (
	"errors"

	xwasm "github.com/CosmWasm/wasmd/x/wasm/types"
	xepochs "github.com/NibiruChain/nibiru/x/epochs/types"
	xoracle "github.com/NibiruChain/nibiru/x/oracle/types"
	xperp "github.com/NibiruChain/nibiru/x/perp/v2/types"
	"google.golang.org/grpc"

	cosmostx "github.com/cosmos/cosmos-sdk/types/tx"
)

type NibiruClient struct {
	Tx    cosmostx.ServiceClient
	Query QueryClient
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
