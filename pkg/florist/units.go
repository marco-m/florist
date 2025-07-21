package florist

import (
	"fmt"
	"strings"
	"time"
)

// HumanDuration is like time.Duration.String(), but instead of stopping at hours (h),
// it also converts days (d) and years (y) with the simple rule that 1d = 24h, 1y = 365d.
// The calculation is simplistic and imprecise (doesn't consider leap years or similar),
// but good enough to give the correct "human feeling"
//
// Largest time representable by time.Duration is 2_540_400 h.
// This implementation is really low performance compared to time.Duration.String()
// but is simple to reason about.
func HumanDuration(d time.Duration) string {
	seconds := int(d / time.Second)
	minutes := seconds / 60
	hours := minutes / 60
	days := hours / 24
	years := days / 365

	seconds = seconds - (60 * minutes)
	minutes = minutes - (60 * hours)
	hours = hours - (24 * days)
	days = days - (365 * years)

	var bld strings.Builder

	if years > 0 {
		fmt.Fprintf(&bld, "%dy", years)
	}
	if days > 0 {
		fmt.Fprintf(&bld, "%dd", days)
	}
	if hours > 0 {
		fmt.Fprintf(&bld, "%dh", hours)
	}
	if minutes > 0 {
		fmt.Fprintf(&bld, "%dm", minutes)
	}
	fmt.Fprintf(&bld, "%ds", seconds)

	return bld.String()
}
