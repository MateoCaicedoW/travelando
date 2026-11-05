package reservations

import (
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiries, err := svc.Inquiries()
	if err != nil {
		slog.Error("loading admin reservations", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	page := system.Layout{Title: "Reservas | Travelando", Portal: system.PortalAdmin, Yield: reservationsPageEl(reservationEntries(inquiries))}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func reservationEntries(inquiries []models.Inquiry) []models.Inquiry {
	reservations := make([]models.Inquiry, 0, len(inquiries))
	for _, inquiry := range inquiries {
		if inquiry.Status == models.InquiryPending || inquiry.Status == models.InquiryReserved || inquiry.Status == models.InquiryCancelled {
			reservations = append(reservations, inquiry)
		}
	}
	return reservations
}
