package profile

import (
	"errors"
	"log/slog"
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Update(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	r.Form = r.PostForm
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	ownerID := auth.OwnerIDFromSession(r)
	current, err := svc.OwnerByID(ownerID)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading owner profile for update", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	state := &ownerProfileState{Errors: ownerProfileErrors(r), Values: r.PostForm}
	if state.Errors.HasAny() {
		renderOwnerProfile(w, r, *current, state, http.StatusUnprocessableEntity)
		return
	}

	newPIN := r.FormValue("new_pin")
	if newPIN != "" {
		authenticated, err := svc.AuthenticateOwner(current.Cedula, r.FormValue("current_pin"))
		if errors.Is(err, services.ErrUnauthorized) || err == nil && authenticated.ID != ownerID {
			addOwnerProfileError(state, "current_pin", "El PIN actual es incorrecto.")
			renderOwnerProfile(w, r, *current, state, http.StatusUnprocessableEntity)
			return
		}
		if err != nil {
			slog.Error("verifying owner PIN for profile update", "error", err)
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
	}

	updated := models.Owner{ID: ownerID, Name: r.FormValue("name"), Cedula: r.FormValue("cedula"), WhatsApp: r.FormValue("whatsapp")}
	if err := svc.UpdateOwner(&updated, newPIN); err != nil {
		switch {
		case errors.Is(err, services.ErrConflict):
			addOwnerProfileError(state, "cedula", "La cédula o el WhatsApp ya están registrados.")
		case errors.Is(err, services.ErrInvalidInput):
			addOwnerProfileError(state, "new_pin", "Revisa los datos ingresados.")
		case errors.Is(err, services.ErrNotFound):
			http.NotFound(w, r)
			return
		default:
			slog.Error("updating owner profile", "error", err)
			server.Error(w, err, http.StatusInternalServerError)
			return
		}
		renderOwnerProfile(w, r, *current, state, http.StatusUnprocessableEntity)
		return
	}

	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Perfil actualizado."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	if updated.Cedula != current.Cedula || updated.WhatsApp != current.WhatsApp || newPIN != "" {
		auth.ClearOwnerSession(r)
		helpers.Redirect(w, r, "/owner/login")
		return
	}
	helpers.Redirect(w, r, "/owner/profile")
}
