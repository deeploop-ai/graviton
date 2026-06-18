package interceptor

import "strings"

func apiKeyScopeAllowed(fullMethod string, scopes []string) bool {
	if len(scopes) == 0 {
		return true
	}
	for _, s := range scopes {
		if s == "*" || s == "all" {
			return true
		}
	}
	resource := apiKeyScopeResource(fullMethod)
	if resource == "" {
		return true
	}
	for _, s := range scopes {
		if s == resource || strings.HasPrefix(s, resource+".") {
			return true
		}
	}
	return false
}

func apiKeyScopeResource(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) < 2 {
		return ""
	}
	svc := parts[len(parts)-2]
	switch {
	case strings.Contains(svc, "Projects"):
		return "projects"
	case strings.Contains(svc, "APIKeys"):
		return "apikeys"
	case strings.Contains(svc, "Users"):
		return "users"
	case strings.Contains(svc, "Teams"):
		return "teams"
	case strings.Contains(svc, "Storage"):
		return "storage"
	case strings.Contains(svc, "Databases"):
		return "databases"
	case strings.Contains(svc, "Health"):
		return "health"
	default:
		return ""
	}
}
