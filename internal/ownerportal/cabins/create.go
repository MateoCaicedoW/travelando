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

func Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, admincabins.MaxRequestBody)
	if err := r.ParseMultipartForm(admincabins.MaxFormMemory); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
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
	cabin, fieldErrors := admincabins.OwnerCabinFromForm(r, 0, ownerID)
	cabinAmenities, amenityErrors := admincabins.CabinAmenitiesFromForm(r, amenityTypes)
	cabin.Amenities = cabinAmenities
	fieldErrors = admincabins.MergeCabinFormErrors(fieldErrors, amenityErrors)
	state := &ownerCabinFormState{Errors: fieldErrors, Values: r.Form, AmenityTypes: amenityTypes}

	files := r.MultipartForm.File["photos"]
	if len(files) > admincabins.MaxPhotos {
		addOwnerCabinFormError(state, "photos", "Solo se pueden subir 5 fotos a la vez.")
		ownerCabinFormError(w, r, page, cabin, state, false)
		return
	}
	for _, fh := range files {
		if fh.Size > admincabins.MaxPhotoSize {
			addOwnerCabinFormError(state, "photos", "Cada foto debe pesar máximo 8MB.")
			ownerCabinFormError(w, r, page, cabin, state, false)
			return
		}
	}
	if fieldErrors.HasAny() {
		ownerCabinFormError(w, r, page, cabin, state, false)
		return
	}

	cabin.Active = true
	if err := set.Cabins.CreateCabinForOwner(ownerID, &cabin); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			addOwnerCabinFormError(state, "name", "Revisa los datos de la cabaña.")
			ownerCabinFormError(w, r, page, cabin, state, false)
			return
		}
		slog.Error("owner creating cabin", "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}

	// Síncrono: los temporales del multipart siguen vivos hasta que el handler retorne.
	if len(files) > 0 {
		if err := admincabins.SaveCabinPhotos(r.Context(), set.Cabins, cabin.ID, files); err != nil {
			slog.Error("owner saving cabin photos", "owner_id", ownerID, "cabin", cabin.ID, "error", err)
			ownerCabinSuccess(w, r, page, "Cabaña agregada, pero algunas fotos no se pudieron subir. Puedes agregarlas al editarla.")
			return
		}
	}
	ownerCabinSuccess(w, r, page, "Cabaña agregada y publicada.")
}
