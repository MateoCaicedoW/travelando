package ownerportal

import (
	"net/http"

	"travelando/internal/reservations"
)

func ShowReservationConfirmation(w http.ResponseWriter, r *http.Request) {
	reservations.ShowConfirmOwner(w, r, ownerIDFromSession(r))
}

func ConfirmReservation(w http.ResponseWriter, r *http.Request) {
	reservations.ConfirmOwner(w, r, ownerIDFromSession(r))
}
