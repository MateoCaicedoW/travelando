package reservations

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func ShowConfirmAdmin(w http.ResponseWriter, r *http.Request) {
	showConfirm(w, r, 0)
}

func ShowConfirmOwner(w http.ResponseWriter, r *http.Request, ownerID int64) {
	showConfirm(w, r, ownerID)
}

func showConfirm(w http.ResponseWriter, r *http.Request, ownerID int64) {
	id, err := reservationID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Inquiries
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	inquiry, err := loadReservation(svc, ownerID, id)
	if errors.Is(err, services.ErrNotFound) || err == nil && inquiry.Status != models.InquiryPending {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading pending inquiry for confirmation", "inquiry_id", id, "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	owner := ownerID > 0
	if err := confirmReservationDialog(*inquiry, owner, strconv.FormatInt(inquiry.Estimated, 10), "").Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
