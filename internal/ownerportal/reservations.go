package ownerportal

import (
	"net/http"

	"travelando/internal/reservations"
)

func EditReservation(w http.ResponseWriter, r *http.Request) {
	reservations.EditOwner(w, r, ownerIDFromSession(r))
}

func UpdateReservation(w http.ResponseWriter, r *http.Request) {
	reservations.UpdateOwner(w, r, ownerIDFromSession(r))
}
