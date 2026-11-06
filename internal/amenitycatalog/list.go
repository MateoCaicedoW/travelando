package amenitycatalog

import "net/http"

func Index(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool) {
	if !admin && ownerID < 1 {
		http.NotFound(w, r)
		return
	}
	renderManager(w, r, ownerID, admin, nil, http.StatusOK)
}
