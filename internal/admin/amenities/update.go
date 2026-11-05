package amenities

import (
	"net/http"

	catalog "travelando/internal/amenitycatalog"
)

func Update(w http.ResponseWriter, r *http.Request) {
	catalog.Update(w, r, 0, true)
}
