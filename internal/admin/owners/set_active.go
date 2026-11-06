package owners

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func SetActive(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := ownerIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	activeValue := r.FormValue("active")
	if activeValue != "0" && activeValue != "1" {
		http.Error(w, "Estado inválido", http.StatusBadRequest)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	active := activeValue == "1"
	if err := svc.SetOwnerActive(id, active); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("updating owner active state", "owner_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	message := "Dueño activado."
	if !active {
		message = "Dueño desactivado; cabañas se mantienen guardadas y dejan de mostrarse."
	}
	if hxhttp.IsRequest(r.Header) {
		renderOwnersPage(w, r, svc, 0, nil, http.StatusOK, helpers.FlashSuccess, message)
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, "/admin/owners")
}
