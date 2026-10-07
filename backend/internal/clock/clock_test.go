package clock

import (
	"testing"
	"time"
)

func at(t *testing.T, utc string) {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, utc)
	if err != nil {
		t.Fatal(err)
	}
	prev := Now
	Now = func() time.Time { return ts }
	t.Cleanup(func() { Now = prev })
}

// The day turns at Bucharest midnight: 21:00 UTC in summer (EEST, UTC+3),
// 22:00 UTC in winter (EET, UTC+2), and the clock-change days in between.
func TestTodayTurnsAtBucharestMidnight(t *testing.T) {
	for _, c := range []struct{ utc, want string }{
		{"2026-07-15T20:59:59Z", "2026-07-15"}, // summer, 23:59:59 in Bucharest
		{"2026-07-15T21:00:00Z", "2026-07-16"}, // summer, midnight in Bucharest
		{"2026-07-15T22:30:00Z", "2026-07-16"}, // summer: still the 15th in UTC
		{"2026-01-15T21:59:59Z", "2026-01-15"}, // winter, 23:59:59 in Bucharest
		{"2026-01-15T22:00:00Z", "2026-01-16"}, // winter, midnight in Bucharest
		{"2026-03-28T21:59:59Z", "2026-03-28"}, // the night before the spring change (still UTC+2)
		{"2026-03-28T22:00:00Z", "2026-03-29"}, // spring change day starts
		{"2026-03-29T00:59:59Z", "2026-03-29"}, // 02:59:59 EET, a second before the jump
		{"2026-03-29T01:00:00Z", "2026-03-29"}, // 04:00 EEST
		{"2026-03-29T20:59:59Z", "2026-03-29"}, // the short day's last second
		{"2026-03-29T21:00:00Z", "2026-03-30"}, // the next day starts at 21:00 UTC now
		{"2026-10-24T20:59:59Z", "2026-10-24"}, // the night before the autumn change (UTC+3)
		{"2026-10-24T21:00:00Z", "2026-10-25"}, // autumn change day starts
		{"2026-10-25T21:59:59Z", "2026-10-25"}, // the long day's last second
		{"2026-10-25T22:00:00Z", "2026-10-26"}, // the next day starts at 22:00 UTC now
	} {
		at(t, c.utc)
		if got := Today(); got != c.want {
			t.Errorf("at %s: Today() = %s, want %s", c.utc, got, c.want)
		}
	}
}

func TestZoneIsBucharest(t *testing.T) {
	if Zone.String() != "Europe/Bucharest" {
		t.Fatalf("Zone = %s", Zone)
	}
}
