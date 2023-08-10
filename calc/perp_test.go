package calc_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/NibiruChain/nibiru/x/common/testutil"
	// perpkeeper "github.com/NibiruChain/nibiru/x/perp/v2/keeper"
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
	priceMult := int64(2)
	testAmm := perptypes.AMM{
		Pair:            "eth:usd",
		BaseReserve:     sdk.NewDec(1_000_000),
		QuoteReserve:    sdk.NewDec(420_000_000),
		SqrtDepth:       sdk.NewDec(420_000_000),
		PriceMultiplier: sdk.NewDec(priceMult),
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
			name: "base case: 1x lev, long, mr 1.1",
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
			outPosNot:       sdk.NewDec(420_000 * priceMult),
			outPnl:          sdk.NewDec(42_000 * priceMult),
		},
		{
			name: "base case: 1x lev, long, mr 1.2",
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
			outPosNot:       sdk.NewDec(420_000 * priceMult),
			outPnl:          sdk.NewDec(84_000 * priceMult),
		},
		{
			name: "base case: 1x lev, short, mr 0.5",
			pos: perptypes.Position{
				TraderAddress:                   trader.String(),
				Pair:                            testAmm.Pair,
				Size_:                           sdk.NewDec(-700),
				Margin:                          sdk.NewDec(420_000),
				OpenNotional:                    sdk.NewDec(420_000),
				LatestCumulativePremiumFraction: sdk.NewDec(0),
				LastUpdatedBlockNumber:          0,
			},
			marketLatestCpf: sdk.NewDec(0),
			marginRatio:     sdk.MustNewDecFromStr("0.42757"),
			newMarginRatio:  sdk.MustNewDecFromStr("0.85"),
			outPosNot:       sdk.NewDec(420_000 * priceMult * 100 / 143),
			outPnl:          sdk.NewDec(84_000 * priceMult * 143 / 100),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fmt.Println("tc: ", tc.name)
			marginRatio, posNot, err := calc.CalcMarginRatioFromAmm(
				tc.pos, testAmm, tc.marketLatestCpf,
			)
			fmt.Printf("marginRatio: %v\n", marginRatio)
			fmt.Printf("posNot: %v\n", posNot)
			require.NoError(t, err)

			// perpkeeper.Keeper.MultiLiquidate(hh)

			pnl, marginRatio, posNot, err := calc.PnlForNewMarginRatio(
				tc.newMarginRatio, tc.pos, testAmm, tc.marketLatestCpf,
			)
			fmt.Printf("tc.newMarginRatio: %v\n", tc.newMarginRatio)
			fmt.Printf("pnl: %v\n", pnl)
			fmt.Printf("marginRatio: %v\n", marginRatio)
			fmt.Printf("posNot: %v\n", posNot)

			fmt.Printf("testAmm.MarkPrice(): %v\n", testAmm.MarkPrice())
			fmt.Println("")

			if tc.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			DecEqualWithinPct(t, 4, tc.marginRatio, marginRatio)
			DecEqualWithinPct(t, 4, tc.outPnl, pnl)
			DecEqualWithinPct(t, 4, tc.outPosNot, posNot)
		})
	}
}
