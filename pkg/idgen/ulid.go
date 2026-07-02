package idgen

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

// ULID returns a new lexicographically sortable, time-ordered identifier.
func ULID() ID {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return ID(ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String())
}
