package amenities

import (
	"net/http"

	catalog "travelando/internal/amenitycatalog"
)

func Index(w http.ResponseWriter, r *http.Request) {
	catalog.Index(w, r, 0, true)
}
