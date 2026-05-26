package clock

import "time"

// Fake is a deterministic Clock for tests. Set or advance T to control time.
type Fake struct {
	T time.Time
}

// NewFake returns a Fake initialized at the given instant.
func NewFake(t time.Time) *Fake { return &Fake{T: t} }

// Now returns the current fake time.
func (f *Fake) Now() time.Time { return f.T }

// Advance moves the fake clock forward by d.
func (f *Fake) Advance(d time.Duration) { f.T = f.T.Add(d) }
