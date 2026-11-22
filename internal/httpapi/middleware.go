package httpapi

import (
	"context"
	"net/http"
	"slices"

	"establishment/v1/internal/auth"
)

type loggedInUserKey struct{}

func requireSession(authenticator sessionAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				writeError(w, auth.ErrSessionNotFound)
				return
			}
			user, err := authenticator.Authenticate(r.Context(), cookie.Value)
			if err != nil {
				writeError(w, err)
				return
			}
			ctx := context.WithValue(r.Context(), loggedInUserKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func loggedInUser(ctx context.Context) (auth.User, bool) {
	user, ok := ctx.Value(loggedInUserKey{}).(auth.User)
	return user, ok
}

func allowOrigins(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if slices.Contains(allowedOrigins, origin) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
