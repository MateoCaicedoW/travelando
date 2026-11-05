package cabins

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func saveCabin(w http.ResponseWriter, r *http.Request, operation func(services.CabinService, *models.Cabin) error) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := cabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	cabin, fieldErrors := cabinFromForm(r, id)
	cabin.Active = r.FormValue("active") == "on" || cabin.ID == 0
	if fieldErrors.HasAny() {
		renderCabinForm(w, r, cabin, &cabinFormState{Errors: fieldErrors, Values: r.Form}, http.StatusUnprocessableEntity)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Cabins
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := operation(svc, &cabin); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			if err := helpers.SetFlash(r, helpers.FlashError, "Revisa los datos de la cabaña."); err != nil {
				http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
				return
			}
			renderCabinForm(w, r, cabin, &cabinFormState{Values: r.Form}, http.StatusUnprocessableEntity)
			return
		}
		slog.Error("saving cabin", "error", err)
		if err := helpers.SetFlash(r, helpers.FlashError, "No se pudo guardar la cabaña."); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		renderCabinForm(w, r, cabin, &cabinFormState{Values: r.Form}, http.StatusInternalServerError)
		return
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
	w.WriteHeader(status)
	if helpers.IsHTMX(r) {
		if err := saveErrorEl(cabin, blocks, bookings, owners, state).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	title := "Nueva cabaña | Travelando"
	if cabin.ID > 0 {
		title = "Editar cabaña | Travelando"
	}
	page := system.Layout{Title: title, Portal: system.PortalAdmin, Yield: saveErrorEl(cabin, blocks, bookings, owners, state)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
