package reservations

import (
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/reservations"
)

func ShowConfirm(w http.ResponseWriter, r *http.Request) {
	reservations.ShowConfirmOwner(w, r, auth.OwnerIDFromSession(r))
}
