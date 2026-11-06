package reservations

import (
	"log/slog"
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiries, err := svc.ReservationsForOwner(auth.OwnerIDFromSession(r))
	if err != nil {
		slog.Error("loading owner reservations", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	page := system.Layout{Title: "Reservas | Travelando", Portal: system.PortalOwner, Yield: ownerReservationsPageEl(inquiries)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
