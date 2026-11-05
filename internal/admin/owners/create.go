package owners

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"
)

func Create(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	state := &ownerFormState{Errors: ownerFormErrors(r, true), Values: r.Form}
	if state.Errors.HasAny() {
		renderOwnerCreateError(w, r, svc, state, http.StatusUnprocessableEntity)
		return
	}
	owner, pin := ownerFromForm(r, 0)
	if err := svc.CreateOwner(&owner, pin); err != nil {
		if errors.Is(err, services.ErrConflict) {
			ownerConflictMessage(state)
			renderOwnerCreateError(w, r, svc, state, http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			addOwnerFormError(state, "pin", "Revisa los datos ingresados.")
			renderOwnerCreateError(w, r, svc, state, http.StatusUnprocessableEntity)
			return
		}
		slog.Error("creating owner access", "error", err)
		http.Error(w, "No se pudo crear el acceso", http.StatusInternalServerError)
		return
	}
	renderOwnerCreateSuccess(w, r, svc)
}
