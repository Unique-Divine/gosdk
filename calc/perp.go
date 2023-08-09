package calc

import (
	perpkeeper "github.com/NibiruChain/nibiru/x/perp/v2/keeper"
	perptypes "github.com/NibiruChain/nibiru/x/perp/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func CalcMarginRatioFromAmm(
	pos perptypes.Position,
	amm perptypes.AMM,
	marketLatestCpf sdk.Dec,
) (marginRatio sdk.Dec, posNot sdk.Dec, err error) {
	posNot, err = perpkeeper.PositionNotionalSpot(amm, pos)
	if err != nil {
		return marginRatio, posNot, err
	}

	marginRatio = perpkeeper.MarginRatio(pos, posNot, marketLatestCpf)
	return marginRatio, posNot, err
}

func CalcMarginRatioFromTwap(
	pos perptypes.Position,
	posNotFromTwap sdk.Dec,
	marketLatestCpf sdk.Dec,
) sdk.Dec {
	return perpkeeper.MarginRatio(pos, posNotFromTwap, marketLatestCpf)
}

// PnlForNewMarginRatio: PnL required to bring about the new margin ratio
func PnlForNewMarginRatio(
	newMarginRatio sdk.Dec, pos perptypes.Position, amm perptypes.AMM,
	marketLatestCpf sdk.Dec,
) (pnl sdk.Dec, marginRatio sdk.Dec, posNot sdk.Dec, err error) {
	marginRatio, posNot, err = CalcMarginRatioFromAmm(pos, amm, marketLatestCpf)
	if err != nil {
		return pnl, marginRatio, posNot, err
	}

	backingMargin := marginRatio.Mul(posNot)
	newBackingMargin := newMarginRatio.Mul(posNot)

	return newBackingMargin.Sub(backingMargin), marginRatio, posNot, err
}

func DirectionMult(posSize sdk.Dec) (dirMult sdk.Int) {
	if posSize.IsNegative() {
		dirMult = sdk.NewInt(1)
	} else {
		dirMult = sdk.NewInt(-1)
	}
	return dirMult
}

func PriceFromMarginRatio(
	newMarginRatio sdk.Dec, pos perptypes.Position, amm perptypes.AMM,
	marketLatestCpf sdk.Dec,
) (newPrice, newPosNot, marginRatio sdk.Dec, err error) {
	// pnlForNew is the PnL required to bring about the new margin ratio
	pnlForNew, marginRatio, posNot, err := PnlForNewMarginRatio(
		newMarginRatio, pos, amm, marketLatestCpf)

	dirMult := DirectionMult(pos.Size_)
	deltaPosNotForPnl := pnlForNew.MulInt(dirMult)
	newPosNot = posNot.Add(deltaPosNotForPnl)
	newPrice = newPosNot.Quo(pos.Size_)
	return newPrice, newPosNot, marginRatio, err
}
