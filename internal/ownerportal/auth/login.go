package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	"go.leapkit.dev/core/server/session"
)

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	state := &loginState{Errors: ownerLoginErrors(r), Values: r.Form}
	if state.Errors.HasAny() {
		renderLoginPage(w, r, state, http.StatusUnprocessableEntity)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owner, err := svc.AuthenticateOwner(r.FormValue("cedula"), r.FormValue("pin"))
	if errors.Is(err, services.ErrUnauthorized) {
		state.Errors["pin"] = append(state.Errors["pin"], errors.New("Cédula o PIN incorrecto. Si fallas cinco veces, espera 15 minutos."))
		renderLoginPage(w, r, state, http.StatusUnauthorized)
		return
	}
	if err != nil {
		slog.Error("authenticating cabin owner", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	sess := session.FromCtx(r.Context())
	if sess == nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	sess.Values[ownerSessionKey] = ownerSessionValue(owner.ID, owner.SessionVersion)
	helpers.Redirect(w, r, "/owner/")
}
