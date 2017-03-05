package echo

import "time"

// Now returns a time message from the current Unix timestamp
func Now() *Time {
	return &Time{Nanoseconds: time.Now().UnixNano()}
}

// Parse a Unix timestamp from an echo.Time message.
func (ts *Time) Parse() time.Time {
	if ts != nil {
		secs := ts.Seconds
		nsecs := ts.Nanoseconds
		return time.Unix(secs, nsecs)
	}
	return time.Time{}
}
