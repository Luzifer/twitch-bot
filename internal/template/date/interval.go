package date

import (
	"fmt"
	"strings"
	"time"
)

type (
	// Interval represents a human-minded diff of two dates which is in
	// no way interchangeable with a time.Duration: The DateInterval
	// takes each date component and subtracts them. This causes the
	// 03/25 to be exactly one month distant from 02/25 even though the
	// distance would be different than with 03/25 and 04/25 which is
	// also exactly one month.
	Interval struct {
		Years   int
		Months  int
		Days    int
		Hours   int
		Minutes int
		Seconds int
	}
)

// NewInterval creates an Interval from two given dates
func NewInterval(a, b time.Time) (i Interval) {
	var l, u time.Time
	if a.Before(b) {
		l, u = a.UTC(), b.UTC()
	} else {
		l, u = b.UTC(), a.UTC()
	}

	i.Years = u.Year() - l.Year()
	i.Months = int(u.Month() - l.Month())
	i.Days = u.Day() - l.Day()
	i.Hours = u.Hour() - l.Hour()
	i.Minutes = u.Minute() - l.Minute()
	i.Seconds = u.Second() - l.Second()

	if i.Seconds < 0 {
		i.Minutes, i.Seconds = i.Minutes-1, i.Seconds+60 //nolint:mnd
	}

	if i.Minutes < 0 {
		i.Hours, i.Minutes = i.Hours-1, i.Minutes+60 //nolint:mnd
	}

	if i.Hours < 0 {
		i.Days, i.Hours = i.Days-1, i.Hours+24 //nolint:mnd
	}

	if i.Days < 0 {
		// oh boi.
		i.Months, i.Days = i.Months-1, daysInMonth(u.Year(), int(u.Month())-1)+i.Days
	}

	if i.Months < 0 {
		i.Years, i.Months = i.Years-1, i.Months+12 //nolint:mnd
	}

	return i
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.Local).Add(-time.Second).Day()
}

// Format takes a template string analog to a strftime string and formats
// the Interval accordingly:
//
// %Y / %y = Years with / without leading digit to 2 places
// %M / %m = Months with / without leading digit to 2 places
// %D / %d = Days with / without leading digit to 2 places
// %H / %h = Hours with / without leading digit to 2 places
// %I / %i = Minutes with / without leading digit to 2 places
// %S / %s = Seconds with / without leading digit to 2 places
func (i Interval) Format(tplString string) string {
	return strings.NewReplacer(
		"%Y", fmt.Sprintf("%02d", i.Years),
		"%y", fmt.Sprintf("%d", i.Years),
		"%M", fmt.Sprintf("%02d", i.Months),
		"%m", fmt.Sprintf("%d", i.Months),
		"%D", fmt.Sprintf("%02d", i.Days),
		"%d", fmt.Sprintf("%d", i.Days),
		"%H", fmt.Sprintf("%02d", i.Hours),
		"%h", fmt.Sprintf("%d", i.Hours),
		"%I", fmt.Sprintf("%02d", i.Minutes),
		"%i", fmt.Sprintf("%d", i.Minutes),
		"%S", fmt.Sprintf("%02d", i.Seconds),
		"%s", fmt.Sprintf("%d", i.Seconds),
	).Replace(tplString)
}
