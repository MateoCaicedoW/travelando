package cabins

import (
	"net/http"

	admincabins "travelando/internal/admin/cabins"
)

func New(w http.ResponseWriter, r *http.Request) {
	page := ownerCabinPageParam(r)
	cabin := admincabins.DefaultNewCabin()
	ownerCabinModalResponse(w, r, cabin, nil, nil, false, page, http.StatusOK)
}
