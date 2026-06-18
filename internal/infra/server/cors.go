package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/deeploop-ai/fleet/internal/pkg/config"
)

func CORSMiddleware(cfg *config.Http_Cors) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if isOriginAllowed(cfg.GetAllowOrigins(), origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if len(cfg.GetAllowMethods()) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(cfg.GetAllowMethods(), ", "))
			}
			if len(cfg.GetAllowHeaders()) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(cfg.GetAllowHeaders(), ", "))
			}
			if cfg.GetAllowCredentials() {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			if cfg.GetMaxAge() > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(cfg.GetMaxAge())))
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isOriginAllowed(allowed []string, origin string) bool {
	for _, o := range allowed {
		if o == "*" || strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}
