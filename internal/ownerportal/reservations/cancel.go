package reservations

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"
	"maragu.dev/gomponents/html"
)

func Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	ownerID := auth.OwnerIDFromSession(r)
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := svc.CancelReservationForOwner(ownerID, id); errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if errors.Is(err, services.ErrInvalidInput) {
		http.Error(w, "La solicitud o reserva ya no está activa.", http.StatusConflict)
		return
	} else if err != nil {
		slog.Error("cancelling owner reservation", "inquiry_id", id, "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if hxhttp.IsRequest(r.Header) {
		inquiries, err := svc.ReservationsForOwner(ownerID)
		if err != nil {
			slog.Error("reloading owner reservations after cancellation", "owner_id", ownerID, "error", err)
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
		response := Group{
			ownerReservationsContentEl(inquiries),
			html.Div(html.ID("modal"), hx.SwapOOB("delete")),
			html.Div(html.ID("global-flash-container"), hx.SwapOOB("true"), components.FlashToast("Cancelación registrada. Las fechas ya no están bloqueadas.", helpers.FlashSuccess)),
		}
		if err := response.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Cancelación registrada. Las fechas ya no están bloqueadas."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, "/owner/reservations")
}
