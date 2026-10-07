// Package clock is the one place this backend decides what day it is
// (lamsza WAYS_OF_WORKING R19): every daily or "today" decision uses the
// network's time zone, Europe/Bucharest, explicitly, never UTC, the database's
// session zone or the visitor's clock.
package clock

import (
	"time"

	// The zone database is compiled in, so loading the zone cannot fail and
	// nothing falls back to UTC on a machine without tzdata.
	_ "time/tzdata"
)

// ZoneName is the network's time zone: Székelyföld is in Romania.
const ZoneName = "Europe/Bucharest"

// Zone is ZoneName, loaded once.
var Zone = mustLoad()

// Now is the current instant. Tests replace it to stand at an edge time
// (midnight, the clock changes) and restore it afterwards.
var Now = time.Now

func mustLoad() *time.Location {
	loc, err := time.LoadLocation(ZoneName)
	if err != nil {
		panic("clock: " + err.Error())
	}
	return loc
}

// Today is the current calendar day in Bucharest, as YYYY-MM-DD.
func Today() string {
	return Now().In(Zone).Format("2006-01-02")
}
