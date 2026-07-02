package idgen

import "errors"

const (
	RandomCharsetNumeric      = "numeric"
	RandomCharsetAlphanumeric = "alphanumeric"
)

// RandomConfig controls fixed-length random identifiers (reserved for future use).
type RandomConfig struct {
	Length  int
	Charset string
}

// ErrRandomStrategyNotImplemented indicates the random ID strategy is configured but not yet implemented.
var ErrRandomStrategyNotImplemented = errors.New("random id strategy is not implemented")
