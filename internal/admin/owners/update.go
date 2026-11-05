package owners

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"
)

func Update(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := ownerIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owners, err := svc.Owners()
	if err != nil {
		slog.Error("loading owner for update", "error", err)
		http.Error(w, "No se pudo cargar el dueño", http.StatusInternalServerError)
		return
	}
	current, found := ownerByID(owners, id)
	if !found {
		http.NotFound(w, r)
		return
	}
	state := &ownerFormState{Errors: ownerFormErrors(r, !current.HasPIN), Values: r.Form}
	if state.Errors.HasAny() {
		renderOwnersPage(w, r, svc, id, state, http.StatusUnprocessableEntity, "", "")
		return
	}
	owner, pin := ownerFromForm(r, id)
	if err := svc.UpdateOwner(&owner, pin); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, services.ErrConflict) {
			ownerConflictMessage(state)
			renderOwnersPage(w, r, svc, id, state, http.StatusUnprocessableEntity, "", "")
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			addOwnerFormError(state, "pin", "Revisa los datos ingresados.")
			renderOwnersPage(w, r, svc, id, state, http.StatusUnprocessableEntity, "", "")
			return
		}
		slog.Error("updating owner access", "error", err)
		http.Error(w, "No se pudo actualizar el acceso", http.StatusInternalServerError)
		return
	}
	if helpers.IsHTMX(r) {
		renderOwnersPage(w, r, svc, 0, nil, http.StatusOK, helpers.FlashSuccess, "Datos del dueño actualizados.")
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Datos del dueño actualizados."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, "/admin/owners")
}
