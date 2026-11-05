package ownerportal

import (
	"net/http"

	admincabins "travelando/internal/admin/cabins"
	catalog "travelando/internal/amenitycatalog"
)

func AmenityTypes(w http.ResponseWriter, r *http.Request) {
	catalog.Index(w, r, ownerIDFromSession(r), false)
}

func CreateAmenityType(w http.ResponseWriter, r *http.Request) {
	catalog.Create(w, r, ownerIDFromSession(r), false)
}

func UpdateAmenityType(w http.ResponseWriter, r *http.Request) {
	catalog.Update(w, r, ownerIDFromSession(r), false)
}

func DeleteAmenityType(w http.ResponseWriter, r *http.Request) {
	catalog.Delete(w, r, ownerIDFromSession(r), false)
}

func CreateCabinAmenityType(w http.ResponseWriter, r *http.Request) {
	admincabins.CreateCabinAmenityType(w, r, ownerIDFromSession(r), true)
}
