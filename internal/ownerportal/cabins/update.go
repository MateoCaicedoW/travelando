package cabins

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"travelando/internal/ownerportal/auth"

	admincabins "travelando/internal/admin/cabins"
	"travelando/internal/system/services"
	"travelando/internal/system/storage"

	"go.leapkit.dev/core/server"
)

func Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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
	for _, photoURL := range r.Form["removed_photos"] {
		_ = set.Cabins.DeleteCabinPhoto(cabin.ID, photoURL)
		_ = set.Storage.DeletePhoto(strings.TrimPrefix(photoURL, "/uploads/"))
	}
	if r.MultipartForm != nil {
		photoSvc := set.Storage
		for _, fh := range r.MultipartForm.File["photos"] {
			f, err := fh.Open()
			if err != nil {
				continue
			}
			rel, err := photoSvc.SavePhoto(f, fh)
			f.Close()
			if err != nil {
				continue
			}
			_ = set.Cabins.AddCabinPhoto(cabin.ID, storage.PublicURL(rel))
		}
	}
	ownerCabinSuccess(w, r, page, "Cabaña actualizada.")
}
