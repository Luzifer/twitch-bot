package date

import (
	_ "embed"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed tzdata
var tzDataEuropeBerlin []byte

func TestNewInterval(t *testing.T) {
	tz, err := time.LoadLocationFromTZData("Europe/Berlin", tzDataEuropeBerlin)
	require.NoError(t, err)

	for i, tc := range []struct {
		A, B time.Time
		Exp  Interval
	}{
		{
			// Plain and simple: 1 Month
			A:   time.Date(2024, 3, 3, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 2, 3, 0, 0, 0, 0, tz),
			Exp: Interval{0, 1, 0, 0, 0, 0},
		},
		{
			// Plain and simple: 1 Month, reversed
			A:   time.Date(2024, 2, 3, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 3, 3, 0, 0, 0, 0, tz),
			Exp: Interval{0, 1, 0, 0, 0, 0},
		},
		{
			// Plain and simple: 1 Year, 1 Month
			A:   time.Date(2023, 2, 3, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 3, 3, 0, 0, 0, 0, tz),
			Exp: Interval{1, 1, 0, 0, 0, 0},
		},
		{
			// 11 Months, so Year and Month needs to be adjusted
			A:   time.Date(2023, 3, 3, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 2, 3, 0, 0, 0, 0, tz),
			Exp: Interval{0, 11, 0, 0, 0, 0},
		},
		{
			// Plain and simple: 2 Days
			A:   time.Date(2024, 3, 3, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 3, 5, 0, 0, 0, 0, tz),
			Exp: Interval{0, 0, 2, 0, 0, 0},
		},
		{
			// 1 Month and a few days, so Month and Day needs to be adjusted
			A:   time.Date(2024, 3, 25, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 5, 5, 0, 0, 0, 0, tz),
			Exp: Interval{0, 1, 9, 23, 0, 0},
		},
		{
			// 1 Month and a few days, so Month and Day needs to be adjusted
			A:   time.Date(2024, 2, 25, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 4, 5, 0, 0, 0, 0, tz),
			Exp: Interval{0, 1, 10, 23, 0, 0},
		},
		{
			// 1 Month and a few days, so Month and Day needs to be adjusted
			A:   time.Date(2024, 1, 25, 0, 0, 0, 0, tz),
			B:   time.Date(2024, 3, 5, 0, 0, 0, 0, tz),
			Exp: Interval{0, 1, 9, 0, 0, 0},
		},
		{
			// 1 Month and a few days, so Month and Day needs to be adjusted
			A:   time.Date(2023, 1, 25, 0, 0, 0, 0, tz),
			B:   time.Date(2023, 3, 5, 0, 0, 0, 0, tz),
			Exp: Interval{0, 1, 8, 0, 0, 0},
		},
		{
			// 1 Day and a few hours, so Day and Hours needs to be adjusted
			A:   time.Date(2024, 3, 5, 14, 0, 0, 0, tz),
			B:   time.Date(2024, 3, 7, 0, 0, 0, 0, tz),
			Exp: Interval{0, 0, 1, 10, 0, 0},
		},
		{
			// 1 Hour and a few minutes, so Hours and Minutes needs to be adjusted
			A:   time.Date(2024, 3, 5, 14, 25, 0, 0, tz),
			B:   time.Date(2024, 3, 5, 16, 12, 0, 0, tz),
			Exp: Interval{0, 0, 0, 1, 47, 0},
		},
		{
			// 1 Minute and a few seconds, so Minutes and Seconds needs to be adjusted
			A:   time.Date(2024, 3, 5, 14, 25, 13, 0, tz),
			B:   time.Date(2024, 3, 5, 14, 27, 0, 0, tz),
			Exp: Interval{0, 0, 0, 0, 1, 47},
		},
		{
			// Nearly one year but a few seconds, everything needs to be adjusted
			A:   time.Date(2024, 3, 5, 14, 25, 0, 0, tz),
			B:   time.Date(2023, 3, 5, 14, 25, 13, 0, tz),
			Exp: Interval{0, 11, 28, 23, 59, 47},
		},
		{
			// Nearly one year but a few seconds, everything needs to be adjusted
			A:   time.Date(2024, 8, 5, 14, 25, 0, 0, tz),
			B:   time.Date(2023, 8, 5, 14, 25, 13, 0, tz),
			Exp: Interval{0, 11, 30, 23, 59, 47},
		},
		{
			// Nearly one year but a few seconds, everything needs to be adjusted
			A:   time.Date(2024, 7, 5, 14, 25, 0, 0, tz),
			B:   time.Date(2023, 7, 5, 14, 25, 13, 0, tz),
			Exp: Interval{0, 11, 29, 23, 59, 47},
		},
		{
			// Nearly one year but a few seconds, everything needs to be adjusted
			A:   time.Date(2024, 2, 5, 14, 25, 0, 0, tz),
			B:   time.Date(2023, 2, 5, 14, 25, 13, 0, tz),
			Exp: Interval{0, 11, 30, 23, 59, 47},
		},
	} {
		assert.Equal(t,
			tc.Exp,
			NewInterval(tc.A, tc.B),
			fmt.Sprintf("%d: %s -> %s", i, tc.A, tc.B))
	}
}

func TestFormatInterval(t *testing.T) {
	ti := Interval{1, 2, 3, 4, 5, 6}
	assert.Equal(t,
		"01 1 years 02 2 months 03 3 days 04 4 hours 05 5 minutes 06 6 seconds",
		ti.Format("%Y %y years %M %m months %D %d days %H %h hours %I %i minutes %S %s seconds"))
}
