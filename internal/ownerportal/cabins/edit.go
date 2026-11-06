package cabins

import (
	"errors"
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Edit(w http.ResponseWriter, r *http.Request) {
	id, err := ownerCabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	cabin, err := set.Cabins.Cabin(id)
	if errors.Is(err, services.ErrNotFound) || err == nil && cabin.OwnerID != auth.OwnerIDFromSession(r) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	ownerCabinModalResponse(w, r, *cabin, nil, nil, true, ownerCabinPageParam(r), http.StatusOK)
}
