package owners

import (
	"net/http"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Edit(w http.ResponseWriter, r *http.Request) {
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
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if !ownerExists(owners, id) {
		http.NotFound(w, r)
		return
	}
	renderOwnersContent(w, r, owners, id, nil, http.StatusOK, "", "")
}
