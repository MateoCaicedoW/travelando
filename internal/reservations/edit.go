package reservations

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func EditAdmin(w http.ResponseWriter, r *http.Request) {
	edit(w, r, 0)
}

func EditOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	edit(w, r, ownerID)
}

func edit(w http.ResponseWriter, r *http.Request, ownerID int64) {
	id, err := reservationID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Inquiries == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiry, err := loadReservation(set.Inquiries, ownerID, id)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading reservation for editing", "inquiry_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if inquiry.Status == models.InquiryCancelled {
		http.NotFound(w, r)
		return
	}
	render(w, r, *inquiry, ownerID, nil, http.StatusOK)
}
