package ownerportal

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/reservations"
	"travelando/internal/system/components"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func ReservationDetails(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiry, err := svc.InquiryForOwner(ownerIDFromSession(r), id)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading owner reservation details", "inquiry_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	label := "Detalles de reserva"
	if inquiry.Status == "pending" || inquiry.CancelledFromStatus == "pending" {
		label = "Detalles de solicitud"
	}
	if err := components.CustomDialog(components.DialogProps{Label: label, Yield: reservations.OwnerReservationDetailsEl(*inquiry)}).Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
