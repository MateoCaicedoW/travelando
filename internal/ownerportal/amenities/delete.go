package amenities

import (
	"net/http"

	"travelando/internal/amenitycatalog"
	"travelando/internal/ownerportal/auth"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	amenitycatalog.Delete(w, r, auth.OwnerIDFromSession(r), false)
}
