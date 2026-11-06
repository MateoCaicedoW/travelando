package auth

import (
	"net/http"

	"travelando/internal/system/helpers"

	"go.leapkit.dev/core/server/session"
)

func Logout(w http.ResponseWriter, r *http.Request) {
	if sess := session.FromCtx(r.Context()); sess != nil {
		delete(sess.Values, ownerSessionKey)
	}
	helpers.Redirect(w, r, "/owner/login")
}
