package reservations

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Show(w http.ResponseWriter, r *http.Request) {
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
	inquiry, err := svc.AdminReservation(id)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading admin reservation details", "inquiry_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if err := components.CustomDialog(components.DialogProps{Label: "Detalles de reserva", Yield: reservationDetailsEl(*inquiry)}).Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
