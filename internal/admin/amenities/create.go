package amenities

import (
	"net/http"

	catalog "travelando/internal/amenitycatalog"
)

func Create(w http.ResponseWriter, r *http.Request) {
	catalog.Create(w, r, 0, true)
}
