package idgen

import "strings"

const (
	StrategyUUID     = "uuid"
	StrategySnowflake = "snowflake"
	StrategySequence = "sequence"
	StrategyRandom   = "random"
)

// NormalizeStrategy returns a supported strategy name or uuid as fallback.
func NormalizeStrategy(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case StrategySnowflake:
		return StrategySnowflake
	case StrategySequence:
		return StrategySequence
	case StrategyRandom:
		return StrategyRandom
	default:
		return StrategyUUID
	}
}
