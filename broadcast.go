package gonibi

import (
	"context"

	"github.com/NibiruChain/nibiru/x/common"
	"github.com/NibiruChain/nibiru/x/common/denoms"
	cmtrpc "github.com/cometbft/cometbft/rpc/client"
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkclienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktypestx "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"
)

func BroadcastMsgs(
	args BroadcastArgs,
	from sdk.AccAddress,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {

	kring := args.kring
	txCfg := args.txCfg
	gosdk := args.gosdk
	broadcaster := args.Broadcaster
	chainID := args.chainID

	info, err := kring.KeyByAddress(from)
	if err != nil {
		return nil, err
	}

	txBuilder := txCfg.NewTxBuilder()
	err = txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, err
	}

	bondDenom := denoms.NIBI
	txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(bondDenom, sdk.NewInt(1000))))
	txBuilder.SetGasLimit(uint64(2 * common.TO_MICRO))

	nums, err := gosdk.GetAccountNumbers(from.String())
	if err != nil {
		return nil, err
	}

	var accRetriever sdkclient.AccountRetriever = authtypes.AccountRetriever{}
	txFactory := sdkclienttx.Factory{}.
		WithChainID(chainID).
		WithKeybase(kring).
		WithTxConfig(txCfg).
		WithAccountRetriever(accRetriever).
		WithAccountNumber(nums.Number).
		WithSequence(nums.Sequence)

	overwriteSig := true
	err = sdkclienttx.Sign(txFactory, info.Name, txBuilder, overwriteSig)
	if err != nil {
		return nil, err
	}

	txBytes, err := txCfg.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return broadcaster.BroadcastTxSync(txBytes)
}

type Broadcaster interface {
	BroadcastTxSync(txBytes []byte) (*sdk.TxResponse, error)
}

var _ Broadcaster = (*BroadcasterTmRpc)(nil)
var _ Broadcaster = (*BroadcasterGrpc)(nil)

type BroadcasterTmRpc struct {
	RPC cmtrpc.Client
}

func (b BroadcasterTmRpc) BroadcastTxSync(
	txBytes []byte,
) (*sdk.TxResponse, error) {

	respRaw, err := b.RPC.BroadcastTxSync(context.Background(), txBytes)
	if err != nil {
		return nil, err
	}

	return sdk.NewResponseFormatBroadcastTx(respRaw), err
}

type BroadcasterGrpc struct {
	GRPC *grpc.ClientConn
}

func (b BroadcasterGrpc) BroadcastTx(
	txBytes []byte, mode sdktypestx.BroadcastMode,
) (*sdk.TxResponse, error) {
	txClient := sdktypestx.NewServiceClient(b.GRPC)
	respRaw, err := txClient.BroadcastTx(
		context.Background(), &sdktypestx.BroadcastTxRequest{
			TxBytes: txBytes,
			Mode:    mode,
		})
	return respRaw.TxResponse, err
}

func (b BroadcasterGrpc) BroadcastTxSync(
	txBytes []byte,
) (*sdk.TxResponse, error) {
	return b.BroadcastTx(txBytes, sdktypestx.BroadcastMode_BROADCAST_MODE_SYNC)
}

func (b BroadcasterGrpc) BroadcastTxAsync(
	txBytes []byte,
) (*sdk.TxResponse, error) {
	return b.BroadcastTx(txBytes, sdktypestx.BroadcastMode_BROADCAST_MODE_ASYNC)
}

// func GetTxBytes() ([]byte, error) {
// 	return txBytes, err
// }

type BroadcastArgs struct {
	kring       keyring.Keyring
	txCfg       sdkclient.TxConfig
	gosdk       NibiruClient
	clientCtx   sdkclient.Context
	Broadcaster Broadcaster
	rpc         cmtrpc.Client
	chainID     string
}

func (nc *NibiruClient) BroadcastMsgs(
	from sdk.AccAddress,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {
	txConfig := nc.EncCfg.TxConfig
	args := BroadcastArgs{
		kring:       nc.Keyring,
		txCfg:       txConfig,
		gosdk:       *nc,
		Broadcaster: BroadcasterTmRpc{RPC: nc.CometRPC},
		rpc:         nc.CometRPC,
		chainID:     nc.ChainId,
	}
	return BroadcastMsgs(args, from, msgs...)
}

func (nc *NibiruClient) BroadcastMsgsGrpc(
	from sdk.AccAddress,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {
	txConfig := nc.EncCfg.TxConfig
	args := BroadcastArgs{
		kring:       nc.Keyring,
		txCfg:       txConfig,
		gosdk:       *nc,
		Broadcaster: BroadcasterGrpc{GRPC: nc.Querier.ClientConn},
		rpc:         nc.CometRPC,
		chainID:     nc.ChainId,
	}
	return BroadcastMsgs(args, from, msgs...)
}
