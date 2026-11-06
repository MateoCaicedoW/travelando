package reservations

import (
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/reservations"
)

func Edit(w http.ResponseWriter, r *http.Request) {
	reservations.EditOwner(w, r, auth.OwnerIDFromSession(r))
}
