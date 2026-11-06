package profile

import (
	"errors"
	"log/slog"
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owner, err := svc.OwnerByID(auth.OwnerIDFromSession(r))
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading owner profile", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderOwnerProfile(w, r, *owner, nil, http.StatusOK)
}
