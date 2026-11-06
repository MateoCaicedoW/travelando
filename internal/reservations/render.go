package reservations

import (
	"net/http"

	"travelando/internal/system/components"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func render(w http.ResponseWriter, r *http.Request, inquiry models.Inquiry, ownerID int64, state *formState, status int) {
	owner := ownerID > 0
	returnTo := r.URL.Query().Get("return_to")
	if state != nil && state.Values.Get("return_to") != "" {
		returnTo = state.Values.Get("return_to")
	}
	if returnTo != "reservations" {
		returnTo = ""
	}
	if hxhttp.IsRequest(r.Header) && state != nil {
		if status != http.StatusOK {
			w.WriteHeader(status)
		}
		if err := reservationForm(inquiry, state, owner, returnTo).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	label := "Editar reserva"
	if inquiry.Status == models.InquiryPending {
		label = "Editar solicitud"
	}
	dialog := components.CustomDialog(components.DialogProps{Label: label, Yield: reservationForm(inquiry, state, owner, returnTo)})
	if err := dialog.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
