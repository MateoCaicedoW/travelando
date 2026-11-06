package cabins

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

const (
	MaxPhotos      = 5
	MaxPhotoSize   = 8 << 20                            // por foto
	MaxRequestBody = (MaxPhotos * MaxPhotoSize) + 1<<20 // ~41MB total
	MaxFormMemory  = 2 << 20                            // lo demás se va a disco
)

func saveCabin(w http.ResponseWriter, r *http.Request, operation func(services.CabinService, *models.Cabin) error) {
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBody)
	if err := r.ParseMultipartForm(MaxFormMemory); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := cabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	set := services.ServiceSetFromCtx(r.Context())
	svc := set.Cabins
	if svc == nil || set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}

	cabin, fieldErrors := cabinFromForm(r, id)

	var existingPhotos []string
	if cabin.ID > 0 {
		if existing, err := svc.Cabin(cabin.ID); err == nil && existing != nil {
			cabin.Photos = existing.Photos
			existingPhotos = make([]string, 0, len(existing.Photos))
			for _, p := range existing.Photos {
				existingPhotos = append(existingPhotos, p.URL)
			}
		}
	}
	amenityTypes, err := set.Amenities.TypesForOwner(cabin.OwnerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabinAmenities, amenityErrors := CabinAmenitiesFromForm(r, amenityTypes)
	cabin.Amenities = cabinAmenities
	fieldErrors = MergeCabinFormErrors(fieldErrors, amenityErrors)
	cabin.Active = r.FormValue("active") == "on" || cabin.ID == 0
	state := &cabinFormState{Errors: fieldErrors, Values: r.Form, AmenityTypes: amenityTypes}

	files := r.MultipartForm.File["photos"]
	if len(files) > MaxPhotos {
		_ = helpers.SetFlash(r, helpers.FlashError, "Solo se pueden subir 5 fotos a la vez.")
		renderCabinForm(w, r, cabin, state, http.StatusUnprocessableEntity)
		return
	}
	for _, fh := range files {
		if fh.Size > MaxPhotoSize {
			_ = helpers.SetFlash(r, helpers.FlashError, "Cada foto debe pesar máximo 8MB.")
			renderCabinForm(w, r, cabin, state, http.StatusUnprocessableEntity)
			return
		}
	}
	if fieldErrors.HasAny() {
		renderCabinForm(w, r, cabin, state, http.StatusUnprocessableEntity)
		return
	}

	if err := operation(svc, &cabin); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			if err := helpers.SetFlash(r, helpers.FlashError, "Revisa los datos de la cabaña."); err != nil {
				http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
				return
			}
			renderCabinForm(w, r, cabin, state, http.StatusUnprocessableEntity)
			return
		}
		slog.Error("saving cabin", "error", err)
		if err := helpers.SetFlash(r, helpers.FlashError, "No se pudo guardar la cabaña."); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		renderCabinForm(w, r, cabin, state, http.StatusInternalServerError)
		return
	}

	// Síncrono: los archivos temporales del multipart siguen vivos mientras
	// el handler no haya retornado.
	ctx := r.Context()
	photoFailed := false

	if urls := OwnedPhotos(existingPhotos, r.Form["removed_photos"]); len(urls) > 0 {
		if err := DeleteCabinPhotos(ctx, svc, cabin.ID, urls); err != nil {
			slog.Error("deleting cabin photos", "cabin", cabin.ID, "error", err)
			photoFailed = true
		}
	}
	if len(files) > 0 {
		if err := SaveCabinPhotos(ctx, svc, cabin.ID, files); err != nil {
			slog.Error("saving cabin photos", "cabin", cabin.ID, "error", err)
			photoFailed = true
		}
	}
	if photoFailed {
		_ = helpers.SetFlash(r, helpers.FlashError, "La cabaña se guardó, pero algunas fotos no se pudieron procesar.")
	}

	helpers.Redirect(w, r, "/admin/cabins")
}

func renderCabinForm(w http.ResponseWriter, r *http.Request, cabin models.Cabin, state *cabinFormState, status int) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Owners == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owners, err := set.Owners.Owners()
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	blocks, bookings, err := loadCabinFormOptions(set.Blocks, set.Inquiries, cabin.ID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}

	isEdit := cabin.ID > 0
	form := newEl(cabin, owners, state.AmenityTypes, state)
	if isEdit {
		form = editModalEl(cabin, blocks, bookings, owners, state.AmenityTypes, state)
	}
	w.WriteHeader(status)

	if hxhttp.IsRequest(r.Header) {
		if err := form.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	if isEdit {
		dialog := components.CustomDialog(components.DialogProps{Label: "Editar cabaña", Yield: form})
		if err := dialog.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	page := system.Layout{Title: "Nueva cabaña | Travelando", Portal: system.PortalAdmin, Yield: form}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
