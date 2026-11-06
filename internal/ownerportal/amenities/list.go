package amenities

import (
	"net/http"

	"travelando/internal/amenitycatalog"
	"travelando/internal/ownerportal/auth"
)

func Index(w http.ResponseWriter, r *http.Request) {
	amenitycatalog.Index(w, r, auth.OwnerIDFromSession(r), false)
}
