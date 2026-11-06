package cabins

import (
	"net/http"
	"travelando/internal/ownerportal/auth"

	admincabins "travelando/internal/admin/cabins"
)

func CreateCabinAmenityType(w http.ResponseWriter, r *http.Request) {
	admincabins.CreateCabinAmenityType(w, r, auth.OwnerIDFromSession(r), true)
}
