package auth

import (
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/helpers"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func LoginPage(w http.ResponseWriter, r *http.Request) {
	if isAuthenticated(r) {
		helpers.Redirect(w, r, "/admin/")
		return
	}
	renderLoginPage(w, r, http.StatusOK, nil)
}

func renderLoginPage(w http.ResponseWriter, r *http.Request, status int, fieldErrors validate.Errors) {
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	if hxhttp.IsRequest(r.Header) {
		if err := loginPageEl(fieldErrors).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Ingresar | Travelando", Portal: system.PortalBare, Yield: loginPageEl(fieldErrors)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
