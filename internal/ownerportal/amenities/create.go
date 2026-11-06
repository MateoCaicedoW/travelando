package amenities

import (
	"net/http"

	"travelando/internal/amenitycatalog"
	"travelando/internal/ownerportal/auth"
)

func Create(w http.ResponseWriter, r *http.Request) {
	amenitycatalog.Create(w, r, auth.OwnerIDFromSession(r), false)
}
