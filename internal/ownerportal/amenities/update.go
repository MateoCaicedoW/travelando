package amenities

import (
	"net/http"

	"travelando/internal/amenitycatalog"
	"travelando/internal/ownerportal/auth"
)

func Update(w http.ResponseWriter, r *http.Request) {
	amenitycatalog.Update(w, r, auth.OwnerIDFromSession(r), false)
}
