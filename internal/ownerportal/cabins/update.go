package cabins

import (
	"errors"
	"log/slog"
	"net/http"
	"travelando/internal/ownerportal/auth"

	admincabins "travelando/internal/admin/cabins"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Update(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := ownerCabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ownerID := auth.OwnerIDFromSession(r)
	page := ownerCabinPageParam(r)
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil || set.Cabins == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	amenityTypes, err := set.Amenities.TypesForOwner(ownerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabin, fieldErrors := admincabins.OwnerCabinFromForm(r, id, ownerID)
	cabinAmenities, amenityErrors := admincabins.CabinAmenitiesFromForm(r, amenityTypes)
	cabin.Amenities = cabinAmenities
	fieldErrors = admincabins.MergeCabinFormErrors(fieldErrors, amenityErrors)
	cabin.Active = r.FormValue("active") == "on"
	state := &ownerCabinFormState{Errors: fieldErrors, Values: r.Form, AmenityTypes: amenityTypes}
	if fieldErrors.HasAny() {
		ownerCabinFormError(w, r, page, cabin, state, true)
		return
	}
	if err := set.Cabins.UpdateCabinForOwner(ownerID, &cabin); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			addOwnerCabinFormError(state, "name", "Revisa los datos de la cabaña.")
			ownerCabinFormError(w, r, page, cabin, state, true)
			return
		}
		slog.Error("owner updating cabin", "owner_id", ownerID, "cabin_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	ownerCabinSuccess(w, r, page, "Cabaña actualizada.")
}
