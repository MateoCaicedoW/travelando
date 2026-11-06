package reservations

import (
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/reservations"
)

func Confirm(w http.ResponseWriter, r *http.Request) {
	reservations.ConfirmOwner(w, r, auth.OwnerIDFromSession(r))
}
