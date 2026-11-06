package cabins

import (
	"errors"
	"log/slog"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"
	"travelando/internal/system/storage"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func saveCabin(w http.ResponseWriter, r *http.Request, operation func(services.CabinService, *models.Cabin) error) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<20)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := cabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	cabin, fieldErrors := cabinFromForm(r, id)
	svc := services.ServiceSetFromCtx(r.Context()).Cabins
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
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
	// Save uploaded photos, if any.
	if r.MultipartForm != nil {
		photoSvc := services.ServiceSetFromCtx(r.Context()).Storage
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
				_ = svc.AddCabinPhoto(cabin.ID, storage.PublicURL(rel))
			}(fh)
		}
		wg.Wait()
	}
	var wg sync.WaitGroup
	for _, photoURL := range r.Form["removed_photos"] {
		wg.Add(1)
		go func(photoURL string) {
			defer wg.Done()
			_ = svc.DeleteCabinPhoto(cabin.ID, photoURL)
			_ = services.ServiceSetFromCtx(r.Context()).Storage.DeletePhoto(strings.TrimPrefix(photoURL, "/uploads/"))
		}(photoURL)
	}
	wg.Wait()
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
	form := newEl(cabin, owners, state.AmenityTypes, state)
	if cabin.ID > 0 {
		form = editModalEl(cabin, blocks, bookings, owners, state.AmenityTypes, state)
	}
	w.WriteHeader(status)
	if cabin.ID > 0 {
		if hxhttp.IsRequest(r.Header) {
			if err := form.Render(w); err != nil {
				server.Error(w, err, http.StatusInternalServerError)
			}
			return
		}
		dialog := components.CustomDialog(components.DialogProps{Label: "Editar cabaña", Yield: form})
		if err := dialog.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if hxhttp.IsRequest(r.Header) {
		if err := form.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	title := "Nueva cabaña | Travelando"
	if cabin.ID > 0 {
		title = "Editar cabaña | Travelando"
	}
	page := system.Layout{Title: title, Portal: system.PortalAdmin, Yield: form}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
