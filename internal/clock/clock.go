// Package clock abstracts time so that code depending on the current instant
// can be tested deterministically.
package clock

import "time"

// Clock returns the current instant. Implementations must be safe for
// concurrent use.
type Clock interface {
	Now() time.Time
}

// System is a Clock backed by time.Now, normalised to UTC.
type System struct{}

// Now returns the current UTC time.
func (System) Now() time.Time { return time.Now().UTC() }
