package idgen

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	RandomCharsetNumeric      = "numeric"
	RandomCharsetAlphanumeric = "alphanumeric"
)

const numericAlphabet = "0123456789"
const alphanumericAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// RandomConfig controls fixed-length random identifiers.
type RandomConfig struct {
	Length  int
	Charset string
}

func (c RandomConfig) normalized() RandomConfig {
	out := c
	if out.Length <= 0 {
		out.Length = 10
	}
	if out.Length > 32 {
		out.Length = 32
	}
	if out.Charset == "" {
		out.Charset = RandomCharsetNumeric
	}
	return out
}

// RandomString returns a random identifier using the given config.
func RandomString(cfg RandomConfig) (string, error) {
	cfg = cfg.normalized()
	alphabet := numericAlphabet
	if cfg.Charset == RandomCharsetAlphanumeric {
		alphabet = alphanumericAlphabet
	}
	max := big.NewInt(int64(len(alphabet)))
	out := make([]byte, cfg.Length)
	for i := range out {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", fmt.Errorf("random id generation failed: %w", err)
		}
		out[i] = alphabet[n.Int64()]
	}
	return string(out), nil
}
