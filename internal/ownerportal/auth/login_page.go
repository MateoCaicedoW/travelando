package auth

import (
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/helpers"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if ownerIDFromSession(r) > 0 {
		helpers.Redirect(w, r, "/owner/")
		return
	}
	renderLoginPage(w, r, nil, http.StatusOK)
}

func renderLoginPage(w http.ResponseWriter, r *http.Request, state *loginState, status int) {
	if hxhttp.IsRequest(r.Header) {
		if status != http.StatusOK && status != http.StatusUnprocessableEntity {
			status = http.StatusUnprocessableEntity
		}
		w.WriteHeader(status)
		if err := ownerLoginEl(state).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(status)
	page := system.Layout{Title: "Acceso de dueños | Travelando", Portal: system.PortalBare, Yield: ownerLoginEl(state)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
