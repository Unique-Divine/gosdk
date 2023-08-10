package cmdctx

import (
	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkcodec "github.com/cosmos/cosmos-sdk/codec"
	sdkcodectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

type ClientCtxBuilder struct {
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
		WithKeyring(args.Keyring).
		WithHomeDir(args.TmCfgRootDir).
		WithChainID(args.ChainID).
		WithInterfaceRegistry(args.InterfaceRegistry).
		WithCodec(args.Codec).
		WithTxConfig(args.TxConfig).
		WithAccountRetriever(args.AccountRetriever)
}
