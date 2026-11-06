package reservations

import (
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/reservations"
)

func Update(w http.ResponseWriter, r *http.Request) {
	reservations.UpdateOwner(w, r, auth.OwnerIDFromSession(r))
}
