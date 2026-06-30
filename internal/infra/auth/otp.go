package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GenerateOTP returns a numeric one-time password with the given number of digits.
func GenerateOTP(digits int) (string, error) {
	if digits <= 0 || digits > 10 {
		return "", fmt.Errorf("invalid otp digits: %d", digits)
	}
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(digits)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	format := fmt.Sprintf("%%0%dd", digits)
	return fmt.Sprintf(format, n), nil
}

// HashOTP returns a SHA-256 hex digest of the raw OTP code.
func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
