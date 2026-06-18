package crud

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// PageTokenVersion is the current version of page token format
	PageTokenVersion = "v1"

	// PageTokenSeparator separates parts of the page token
	PageTokenSeparator = ":"

	// DefaultTokenTTL is the default time-to-live for page tokens
	DefaultTokenTTL = 24 * time.Hour

	// TokenModeLegacy represents offset-based token mode.
	TokenModeLegacy = "legacy"
	// TokenModeCursor represents cursor-based token mode.
	TokenModeCursor = "cursor"
)

// PageTokenData represents the internal structure of a page token
// Following AIP-158: https://google.aip.dev/158
type PageTokenData struct {
	Version      string            `json:"v"`                       // Token format version
	Mode         string            `json:"mode,omitempty"`          // legacy | cursor
	Offset       int               `json:"offset"`                  // Current offset
	Created      time.Time         `json:"created"`                 // Token creation time
	PageSize     int32             `json:"page_size"`               // Original page size
	OrderBy      string            `json:"order_by,omitempty"`      // OrderBy of original request
	FilterDigest string            `json:"filter_digest,omitempty"` // Canonical filter digest
	Keys         map[string]string `json:"keys,omitempty"`          // Cursor seek keys
	TTL          int64             `json:"ttl,omitempty"`           // Token TTL in seconds
	Checksum     string            `json:"checksum,omitempty"`      // Compatibility checksum
	Sig          string            `json:"sig,omitempty"`           // HMAC-SHA256 signature
}

// EncodePageToken creates a page token from offset
// This is a simplified version following AIP-158
func EncodePageToken(offset int) string {
	data := PageTokenData{
		Version: PageTokenVersion,
		Mode:    TokenModeLegacy,
		Offset:  offset,
		Created: time.Now().UTC(),
	}

	return encodeTokenData(data)
}

// EncodePageTokenWithSize creates a page token with page size info
func EncodePageTokenWithSize(offset int, pageSize int32) string {
	data := PageTokenData{
		Version:  PageTokenVersion,
		Mode:     TokenModeLegacy,
		Offset:   offset,
		Created:  time.Now().UTC(),
		PageSize: pageSize,
	}

	return encodeTokenData(data)
}

// EncodePageTokenFull creates a page token with full metadata
func EncodePageTokenFull(offset int, pageSize int32, checksum string) string {
	data := PageTokenData{
		Version:  PageTokenVersion,
		Mode:     TokenModeLegacy,
		Offset:   offset,
		Created:  time.Now().UTC(),
		PageSize: pageSize,
		Checksum: checksum,
	}

	return encodeTokenData(data)
}

// EncodeCursorPageToken creates a cursor-mode page token with metadata.
func EncodeCursorPageToken(offset int, pageSize int32, orderBy, filterDigest string, keys map[string]string, checksum string) string {
	data := PageTokenData{
		Version:      PageTokenVersion,
		Mode:         TokenModeCursor,
		Offset:       offset,
		Created:      time.Now().UTC(),
		PageSize:     pageSize,
		OrderBy:      strings.TrimSpace(orderBy),
		FilterDigest: strings.TrimSpace(filterDigest),
		Keys:         keys,
		TTL:          int64(DefaultTokenTTL / time.Second),
		Checksum:     strings.TrimSpace(checksum),
	}
	return encodeTokenData(data)
}

// EncodeSignedCursorPageToken creates a signed cursor-mode token.
func EncodeSignedCursorPageToken(offset int, pageSize int32, orderBy, filterDigest string, keys map[string]string, checksum, secret string) (string, error) {
	data := PageTokenData{
		Version:      PageTokenVersion,
		Mode:         TokenModeCursor,
		Offset:       offset,
		Created:      time.Now().UTC(),
		PageSize:     pageSize,
		OrderBy:      strings.TrimSpace(orderBy),
		FilterDigest: strings.TrimSpace(filterDigest),
		Keys:         keys,
		TTL:          int64(DefaultTokenTTL / time.Second),
		Checksum:     strings.TrimSpace(checksum),
	}
	sig, err := SignPageToken(&data, secret)
	if err != nil {
		return "", err
	}
	data.Sig = sig
	return encodeTokenData(data), nil
}

// SignPageToken computes HMAC-SHA256 signature for token payload.
func SignPageToken(data *PageTokenData, secret string) (string, error) {
	if data == nil {
		return "", fmt.Errorf("page token data is nil")
	}
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("page token secret is required")
	}
	canonical := *data
	canonical.Sig = ""
	raw, err := json.Marshal(canonical)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(raw)
	return hex.EncodeToString(mac.Sum(nil)), nil
}

// VerifyPageTokenSignature validates page token signature.
func VerifyPageTokenSignature(data *PageTokenData, secret string) error {
	if data == nil {
		return fmt.Errorf("page token data is nil")
	}
	if data.Sig == "" {
		return fmt.Errorf("page token signature missing")
	}
	expected, err := SignPageToken(data, secret)
	if err != nil {
		return err
	}
	if !hmac.Equal([]byte(expected), []byte(data.Sig)) {
		return fmt.Errorf("invalid page token signature")
	}
	return nil
}

// encodeTokenData encodes token data to base64 string
func encodeTokenData(data PageTokenData) string {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		// Fallback to simple format
		return fmt.Sprintf("%s%s%d", PageTokenVersion, PageTokenSeparator, data.Offset)
	}

	// Base64 encode the JSON
	return base64.URLEncoding.EncodeToString(jsonBytes)
}

// DecodePageToken decodes a page token and returns the offset
// Following AIP-158 standard
func DecodePageToken(token string) (int, error) {
	if token == "" {
		return 0, nil
	}

	// Try to decode as base64 JSON first (new format)
	data, err := decodeTokenData(token)
	if err == nil {
		// Validate token version
		if data.Version != PageTokenVersion {
			return 0, fmt.Errorf("invalid page token version: %s", data.Version)
		}

		// Check token expiration
		ttl := DefaultTokenTTL
		if data.TTL > 0 {
			ttl = time.Duration(data.TTL) * time.Second
		}
		if time.Since(data.Created) > ttl {
			return 0, fmt.Errorf("page token has expired")
		}

		return data.Offset, nil
	}

	// Fallback: try old simple format "version:offset"
	parts := strings.Split(token, PageTokenSeparator)
	if len(parts) == 2 {
		var offset int
		_, err := fmt.Sscanf(token, "%s:%d", &data.Version, &offset)
		if err == nil && data.Version == PageTokenVersion {
			return offset, nil
		}
	}

	return 0, fmt.Errorf("invalid page token format")
}

// decodeTokenData decodes token data from base64 string
func decodeTokenData(token string) (PageTokenData, error) {
	var data PageTokenData

	// Decode base64
	jsonBytes, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return data, fmt.Errorf("failed to decode page token: %w", err)
	}

	// Unmarshal JSON
	err = json.Unmarshal(jsonBytes, &data)
	if err != nil {
		return data, fmt.Errorf("failed to parse page token data: %w", err)
	}

	return data, nil
}

// DecodePageTokenFull decodes a page token and returns full metadata
func DecodePageTokenFull(token string) (*PageTokenData, error) {
	if token == "" {
		return nil, fmt.Errorf("empty page token")
	}

	data, err := decodeTokenData(token)
	if err != nil {
		return nil, err
	}

	// Validate token version
	if data.Version != PageTokenVersion {
		return nil, fmt.Errorf("invalid page token version: %s", data.Version)
	}

	// Check token expiration
	ttl := DefaultTokenTTL
	if data.TTL > 0 {
		ttl = time.Duration(data.TTL) * time.Second
	}
	if time.Since(data.Created) > ttl {
		return nil, fmt.Errorf("page token has expired")
	}
	if data.Mode == "" {
		data.Mode = TokenModeLegacy
	}

	return &data, nil
}

// DecodeSignedPageTokenFull decodes and validates token signature.
func DecodeSignedPageTokenFull(token, secret string) (*PageTokenData, error) {
	data, err := DecodePageTokenFull(token)
	if err != nil {
		return nil, err
	}
	if err := VerifyPageTokenSignature(data, secret); err != nil {
		return nil, err
	}
	return data, nil
}

// DecodeAndValidatePageToken decodes and validates a page token
func DecodeAndValidatePageToken(token string) (*PageTokenInfo, error) {
	data, err := DecodePageTokenFull(token)
	if err != nil {
		return nil, err
	}

	return &PageTokenInfo{
		Offset: data.Offset,
		Token:  data,
	}, nil
}

// ValidatePageToken checks if a page token is valid without decoding it
func ValidatePageToken(token string) error {
	if token == "" {
		return nil
	}

	_, err := DecodePageToken(token)
	return err
}

// IsExpired checks if a page token has expired
func IsExpired(token string) bool {
	data, err := decodeTokenData(token)
	if err != nil {
		return true
	}
	ttl := DefaultTokenTTL
	if data.TTL > 0 {
		ttl = time.Duration(data.TTL) * time.Second
	}
	return time.Since(data.Created) > ttl
}

// GetTokenAge returns the age of a page token
func GetTokenAge(token string) (time.Duration, error) {
	data, err := decodeTokenData(token)
	if err != nil {
		return 0, err
	}

	return time.Since(data.Created), nil
}

// CalculateTotalPages calculates the total number of pages
func CalculateTotalPages(totalCount int, pageSize int32) int {
	if pageSize <= 0 {
		return 0
	}
	if totalCount <= 0 {
		return 0
	}

	pages := totalCount / int(pageSize)
	if totalCount%int(pageSize) > 0 {
		pages++
	}

	return pages
}

// CalculateCurrentPage calculates the current page number (1-indexed)
func CalculateCurrentPage(offset int, pageSize int32) int {
	if pageSize <= 0 {
		return 1
	}
	return (offset / int(pageSize)) + 1
}

// PaginationInfo contains detailed pagination information
type PaginationInfo struct {
	CurrentPage    int   `json:"current_page"`
	PageSize       int32 `json:"page_size"`
	TotalPages     int   `json:"total_pages,omitempty"`
	TotalCount     int   `json:"total_count,omitempty"`
	HasPrevious    bool  `json:"has_previous"`
	HasNext        bool  `json:"has_next"`
	NextOffset     int   `json:"next_offset,omitempty"`
	PreviousOffset int   `json:"previous_offset,omitempty"`
}

// BuildPaginationInfo builds complete pagination info
func BuildPaginationInfo(params ListParams, totalCount int, hasMore bool) PaginationInfo {
	info := PaginationInfo{
		CurrentPage: CalculateCurrentPage(params.Offset, params.PageSize),
		PageSize:    params.PageSize,
		HasPrevious: params.Offset > 0,
		HasNext:     hasMore,
	}

	if totalCount >= 0 {
		info.TotalCount = totalCount
		info.TotalPages = CalculateTotalPages(totalCount, params.PageSize)
	}

	if hasMore {
		info.NextOffset = params.Offset + int(params.PageSize)
	}

	if params.Offset > 0 {
		info.PreviousOffset = max(0, params.Offset-int(params.PageSize))
	}

	return info
}

// BuildPaginationInfoSimple builds simple pagination info without total count
func BuildPaginationInfoSimple(params ListParams, resultCount int, hasMore bool) PaginationInfo {
	info := PaginationInfo{
		CurrentPage: CalculateCurrentPage(params.Offset, params.PageSize),
		PageSize:    params.PageSize,
		HasPrevious: params.Offset > 0,
		HasNext:     hasMore,
	}

	if hasMore {
		info.NextOffset = params.Offset + resultCount
	}

	if params.Offset > 0 {
		info.PreviousOffset = max(0, params.Offset-int(params.PageSize))
	}

	return info
}

// GeneratePreviousPageToken generates a page token for the previous page
func GeneratePreviousPageToken(params ListParams) string {
	if params.Offset <= 0 {
		return ""
	}

	previousOffset := max(0, params.Offset-int(params.PageSize))
	return EncodePageToken(previousOffset)
}

// GeneratePageTokens generates both next and previous page tokens
func GeneratePageTokens(params ListParams, resultCount int, totalCount int) (nextToken, previousToken string) {
	// Next page token
	if params.Offset+resultCount < totalCount {
		nextToken = EncodePageToken(params.Offset + resultCount)
	}

	// Previous page token
	if params.Offset > 0 {
		previousToken = EncodePageToken(max(0, params.Offset-int(params.PageSize)))
	}

	return nextToken, previousToken
}

// TokenChecksum calculates a checksum for page token validation
func TokenChecksum(data string) string {
	// Simple checksum implementation
	// In production, use a proper hash function like SHA256
	const maxLen = 8
	checksum := 0
	for i, c := range data {
		checksum += int(c) * (i + 1)
	}

	result := ""
	for checksum > 0 {
		result = string(rune('A'+(checksum%26))) + result
		checksum = checksum / 26
	}

	for len(result) < maxLen {
		result = "A" + result
	}

	if len(result) > maxLen {
		result = result[:maxLen]
	}

	return result
}

// ValidatePageTokenChecksum validates the checksum in a page token
func ValidatePageTokenChecksum(token string) bool {
	data, err := decodeTokenData(token)
	if err != nil {
		return false
	}

	if data.Checksum == "" {
		// No checksum to validate
		return true
	}

	// In a real implementation, you would calculate the expected checksum
	// based on the original request parameters and compare it
	// For now, just check that it's not empty
	return data.Checksum != ""
}
