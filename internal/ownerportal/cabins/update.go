package cabins

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/ownerportal/auth"

	admincabins "travelando/internal/admin/cabins"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, admincabins.MaxRequestBody)
	if err := r.ParseMultipartForm(admincabins.MaxFormMemory); err != nil {
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

	var existingPhotos []string
	if existing, err := set.Cabins.Cabin(cabin.ID); err == nil && existing != nil {
		cabin.Photos = existing.Photos
		existingPhotos = make([]string, 0, len(existing.Photos))
		for _, p := range existing.Photos {
			existingPhotos = append(existingPhotos, p.URL) // <- ajusta el nombre del campo
		}
	}

	cabinAmenities, amenityErrors := admincabins.CabinAmenitiesFromForm(r, amenityTypes)
	cabin.Amenities = cabinAmenities
	fieldErrors = admincabins.MergeCabinFormErrors(fieldErrors, amenityErrors)
	cabin.Active = r.FormValue("active") == "on"
	state := &ownerCabinFormState{Errors: fieldErrors, Values: r.Form, AmenityTypes: amenityTypes}

	files := r.MultipartForm.File["photos"]
	if len(files) > admincabins.MaxPhotos {
		addOwnerCabinFormError(state, "photos", "Solo se pueden subir 5 fotos a la vez.")
		ownerCabinFormError(w, r, page, cabin, state, true)
		return
	}
	for _, fh := range files {
		if fh.Size > admincabins.MaxPhotoSize {
			addOwnerCabinFormError(state, "photos", "Cada foto debe pesar máximo 8MB.")
			ownerCabinFormError(w, r, page, cabin, state, true)
			return
		}
	}
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

	// Síncrono: los temporales del multipart siguen vivos hasta que el handler retorne.
	ctx := r.Context()
	photoFailed := false

	if urls := admincabins.OwnedPhotos(existingPhotos, r.Form["removed_photos"]); len(urls) > 0 {
		if err := admincabins.DeleteCabinPhotos(ctx, set.Cabins, cabin.ID, urls); err != nil {
			slog.Error("owner deleting cabin photos", "owner_id", ownerID, "cabin", cabin.ID, "error", err)
			photoFailed = true
		}
	}
	if len(files) > 0 {
		if err := admincabins.SaveCabinPhotos(ctx, set.Cabins, cabin.ID, files); err != nil {
			slog.Error("owner saving cabin photos", "owner_id", ownerID, "cabin", cabin.ID, "error", err)
			photoFailed = true
		}
	}
	if photoFailed {
		ownerCabinSuccess(w, r, page, "Cabaña actualizada, pero algunas fotos no se pudieron procesar. Intenta de nuevo.")
		return
	}
	ownerCabinSuccess(w, r, page, "Cabaña actualizada.")
}
