package handlers

import (
	"context"
	"net/http"
	"strings"
)

// context key type to avoid collisions
type ctxKey string

const claimsContextKey ctxKey = "authClaims"

// RequireAuth is a middleware that:
// - Extracts and validates a Bearer JWT from the Authorization header
// - On success, attaches *Claims to the request context
// - On failure, returns 401 with a JSON error
func (h *AuthHandler) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if authz == "" || !strings.HasPrefix(authz, "Bearer ") {
			WriteError(w, http.StatusUnauthorized, "missing or invalid Authorization header")
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if token == "" {
			WriteError(w, http.StatusUnauthorized, "missing bearer token")
			return
		}

		claims, err := ValidateToken(token, h.cfg.Secret, h.cfg.Issuer)
		if err != nil {
			WriteError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuthFunc is a convenience wrapper to use the middleware with http.HandlerFunc directly.
func (h *AuthHandler) RequireAuthFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.RequireAuth(next).ServeHTTP(w, r)
	}
}

// ClaimsFromContext extracts JWT claims from context if present.
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	val := ctx.Value(claimsContextKey)
	if val == nil {
		return nil, false
	}
	claims, ok := val.(*Claims)
	return claims, ok
}

// ClaimsFromRequest is a helper to extract claims directly from *http.Request.
func ClaimsFromRequest(r *http.Request) (*Claims, bool) {
	return ClaimsFromContext(r.Context())
}

// UserIDFromContext is a convenience helper to get the user ID from context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	if claims, ok := ClaimsFromContext(ctx); ok && claims != nil && claims.UserID != "" {
		return claims.UserID, true
	}
	return "", false
}
