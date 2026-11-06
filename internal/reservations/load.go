package reservations

import (
	"errors"
	"net/http"
	"strconv"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func loadReservation(svc services.InquiryService, ownerID, id int64) (*models.Inquiry, error) {
	if ownerID > 0 {
		return svc.InquiryForOwner(ownerID, id)
	}
	return svc.AdminReservation(id)
}

func reservationID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid reservation ID")
	}
	return id, nil
}
