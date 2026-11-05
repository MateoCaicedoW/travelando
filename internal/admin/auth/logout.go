package auth

import (
	"net/http"

	"go.leapkit.dev/core/server/session"
	"travelando/internal/system/helpers"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	sess := session.FromCtx(r.Context())
	if sess != nil {
		delete(sess.Values, sessionKey)
	}
	helpers.Redirect(w, r, "/")
}
