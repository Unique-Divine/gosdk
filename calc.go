package gosdk

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
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
