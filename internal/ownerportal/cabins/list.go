package cabins

import (
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	page := ownerCabinPageParam(r)
	renderOwnerCabins(w, r, page, http.StatusOK, "", "", false)
}
