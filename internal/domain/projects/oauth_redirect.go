package projects

import (
	"net/url"
	"strings"
)

const SettingsKeyOAuthAllowedRedirectURLs = "auth.oauth_allowed_redirect_urls"

// OAuthAllowedRedirectURLs reads configured redirect URL allowlist from project settings.
func OAuthAllowedRedirectURLs(settings map[string]any) []string {
	if settings == nil {
		return nil
	}
	raw, ok := settings[SettingsKeyOAuthAllowedRedirectURLs]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return append([]string(nil), v...)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

// DefaultOAuthRedirectAllowlist returns fallback allowed origins for development.
func DefaultOAuthRedirectAllowlist(publicBaseURL string) []string {
	out := []string{
		"http://localhost",
		"http://127.0.0.1",
		"https://localhost",
		"https://127.0.0.1",
	}
	if origin := urlOrigin(publicBaseURL); origin != "" {
		out = append(out, origin)
	}
	return out
}

// MatchRedirectURL reports whether rawURL matches one of the allowed redirect entries.
func MatchRedirectURL(rawURL string, allowed []string) bool {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	for _, entry := range allowed {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		pattern, err := url.Parse(entry)
		if err != nil || pattern.Scheme == "" || pattern.Host == "" {
			continue
		}
		if !strings.EqualFold(u.Scheme, pattern.Scheme) || !strings.EqualFold(u.Host, pattern.Host) {
			continue
		}
		if pattern.Path == "" || pattern.Path == "/" {
			return true
		}
		if strings.HasPrefix(u.Path, pattern.Path) {
			return true
		}
	}
	return false
}

func urlOrigin(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}
