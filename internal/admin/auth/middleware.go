package auth

import (
	"net/http"

	"go.leapkit.dev/core/server/session"
	"travelando/internal/system/helpers"
)

const sessionKey = "travelando_admin"

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			helpers.Redirect(w, r, "/admin/login")
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		next.ServeHTTP(w, r)
	})
}

func isAuthenticated(r *http.Request) bool {
	sess := session.FromCtx(r.Context())
	if sess == nil {
		return false
	}
	authenticated, _ := sess.Values[sessionKey].(bool)
	return authenticated
}
