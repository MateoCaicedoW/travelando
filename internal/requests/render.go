package requests

import (
	"net/http"
	"net/url"

	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
)

func renderReportPage(w http.ResponseWriter, r *http.Request, status int, title string, inquiry models.Inquiry, token string, values url.Values, fieldErrors validate.Errors) {
	if helpers.IsHTMX(r) && status == http.StatusConflict {
		helpers.Redirect(w, r, "/report/"+token)
		return
	}
	if status != 0 {
		w.WriteHeader(status)
	}
	if helpers.IsHTMX(r) {
		if err := reportPage(inquiry, token, values, fieldErrors).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: title, Yield: reportPage(inquiry, token, values, fieldErrors)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func renderError(w http.ResponseWriter, r *http.Request, status int, message string) {
	if err := helpers.SetFlash(r, helpers.FlashError, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	page := system.Layout{Title: "No se pudo continuar | Travelando", Yield: errorPage()}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
