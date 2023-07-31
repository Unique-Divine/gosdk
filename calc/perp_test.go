package calc_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NibiruChain/nibiru/x/common/testutil"
	perptypes "github.com/NibiruChain/nibiru/x/perp/v2/types"
	"github.com/Unique-Divine/gonibi/calc"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DecEqualWithinPct(t *testing.T, pctErr int64, numA, numB sdk.Dec) {
	pctErrMult := sdk.NewDec(pctErr).QuoInt64(100)
	errorBand := numA.Mul(pctErrMult)
	diff := numA.Sub(numB)

	errLines := []string{
		fmt.Sprintf("want: %s, got: %s", numA, numB),
		fmt.Sprintf("error_band: %s, diff: %s, pct_err: %d", errorBand, diff, pctErr),
	}
	assert.True(t,
		diff.Abs().LT(errorBand.Abs()),
		strings.Join(errLines, "\n"),
	)
}

func TestPnlForNewMarginRatio(t *testing.T) {
	testAmm := perptypes.AMM{
		Pair:            "eth:usd",
		BaseReserve:     sdk.NewDec(1_000_000),
		QuoteReserve:    sdk.NewDec(420_000_000),
		SqrtDepth:       sdk.NewDec(420_000_000),
		PriceMultiplier: sdk.NewDec(2),
		TotalLong:       sdk.ZeroDec(),
		TotalShort:      sdk.ZeroDec(),
	}
	trader := testutil.AccAddress()
	testCases := []struct {
		name            string
		pos             perptypes.Position
		marketLatestCpf sdk.Dec
		marginRatio     sdk.Dec
		newMarginRatio  sdk.Dec
		outPosNot       sdk.Dec
		outPnl          sdk.Dec
		expectErr       bool
	}{
		{
			name: "base case, 1x lev, mr 1.1",
			pos: perptypes.Position{
				TraderAddress:                   trader.String(),
				Pair:                            testAmm.Pair,
				Size_:                           sdk.NewDec(1_000),
				Margin:                          sdk.NewDec(420_000),
				OpenNotional:                    sdk.NewDec(420_000),
				LatestCumulativePremiumFraction: sdk.NewDec(0),
				LastUpdatedBlockNumber:          0,
			},
			marketLatestCpf: sdk.NewDec(0),
			marginRatio:     sdk.NewDec(1),
			newMarginRatio:  sdk.MustNewDecFromStr("1.1"),
			outPosNot:       sdk.NewDec(840_000),
			outPnl:          sdk.NewDec(84_000),
		},
		{
			name: "base case, 1x lev, mr 1.2",
			pos: perptypes.Position{
				TraderAddress:                   trader.String(),
				Pair:                            testAmm.Pair,
				Size_:                           sdk.NewDec(1_000),
				Margin:                          sdk.NewDec(420_000),
				OpenNotional:                    sdk.NewDec(420_000),
				LatestCumulativePremiumFraction: sdk.NewDec(0),
				LastUpdatedBlockNumber:          0,
			},
			marketLatestCpf: sdk.NewDec(0),
			marginRatio:     sdk.NewDec(1),
			newMarginRatio:  sdk.MustNewDecFromStr("1.2"),
			outPosNot:       sdk.NewDec(840_000),
			outPnl:          sdk.NewDec(168_000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pnl, marginRatio, posNot, err := calc.PnlForNewMarginRatio(
				tc.newMarginRatio, tc.pos, testAmm, tc.marketLatestCpf,
			)

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			DecEqualWithinPct(t, 2, tc.outPnl, pnl)
			DecEqualWithinPct(t, 2, tc.outPosNot, posNot)
			DecEqualWithinPct(t, 2, tc.marginRatio, marginRatio)
		})
	}
}
