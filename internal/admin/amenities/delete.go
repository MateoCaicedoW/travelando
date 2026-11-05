package amenities

import (
	"net/http"

	catalog "travelando/internal/amenitycatalog"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	catalog.Delete(w, r, 0, true)
}
