package florist_test

import (
	"testing"
	"time"

	"github.com/marco-m/florist/pkg/florist"
)

func TestHumanDuration(t *testing.T) {
	test := func(d time.Duration, want string) {
		if have := florist.HumanDuration(d); have != want {
			t.Errorf("\nhave: %s\nwant: %s", have, want)
		}
	}

	test(0, "0s")
	test(time.Minute+3*time.Second, "1m3s")
	test(time.Minute+3*time.Second+500*time.Millisecond, "1m3s") // yeah truncated
	test(5*time.Hour+4*time.Minute+3*time.Second, "5h4m3s")
	test(24*time.Hour+5*time.Hour+4*time.Minute+3*time.Second, "1d5h4m3s")
	test(51*24*time.Hour+5*time.Hour+4*time.Minute+3*time.Second, "51d5h4m3s")
	test(365*24*time.Hour+51*24*time.Hour+5*time.Hour+4*time.Minute+3*time.Second, "1y51d5h4m3s")
	test(7*365*24*time.Hour+51*24*time.Hour+5*time.Hour+4*time.Minute+3*time.Second, "7y51d5h4m3s")
}
