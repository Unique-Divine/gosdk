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
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func BroadcastMsgs(
	args BroadcastArgs,
	from sdk.AccAddress,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {

	kring := args.kring
	txCfg := args.txCfg
	gosdk := args.gosdk
	// broadcaster := args.GrpcBroadcaster
	rpc := gosdk.CometRPC
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

	respRaw, err := rpc.BroadcastTxSync(context.Background(), txBytes)
	if err != nil {
		return nil, err
	}

	return sdk.NewResponseFormatBroadcastTx(respRaw), err
}

// func GetTxBytes() ([]byte, error) {
// 	return txBytes, err
// }

type BroadcastArgs struct {
	kring           keyring.Keyring
	txCfg           sdkclient.TxConfig
	gosdk           NibiruClient
	clientCtx       sdkclient.Context
	GrpcBroadcaster GrpcBroadcastClient
	rpc             cmtrpc.Client
	chainID         string
}

func (nc *NibiruClient) BroadcastMsgs(
	from sdk.AccAddress,
	msgs ...sdk.Msg,
) (*sdk.TxResponse, error) {
	txConfig := nc.EncCfg.TxConfig
	args := BroadcastArgs{
		kring:           nc.Keyring,
		txCfg:           txConfig,
		gosdk:           *nc,
		GrpcBroadcaster: nc.Tx,
		rpc:             nc.CometRPC,
		chainID:         nc.ChainId,
	}
	return BroadcastMsgs(args, from, msgs...)
}
