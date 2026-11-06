package cabins

import (
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"
	"sync"

	"travelando/internal/ownerportal/auth"

	admincabins "travelando/internal/admin/cabins"
	"travelando/internal/system/services"
	"travelando/internal/system/storage"

	"go.leapkit.dev/core/server"
)

func Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
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
	if r.MultipartForm != nil {
		photoSvc := set.Storage
		var wg sync.WaitGroup
		for _, fh := range r.MultipartForm.File["photos"] {
			wg.Add(1)
			go func(fh *multipart.FileHeader) {
				defer wg.Done()
				f, err := fh.Open()
				if err != nil {
					return
				}
				defer f.Close()
				rel, err := photoSvc.SavePhoto(f, fh)
				if err != nil {
					return
				}
				_ = set.Cabins.AddCabinPhoto(cabin.ID, storage.PublicURL(rel))
			}(fh)
		}
		wg.Wait()
	}
	ownerCabinSuccess(w, r, page, "Cabaña agregada y publicada.")
}
