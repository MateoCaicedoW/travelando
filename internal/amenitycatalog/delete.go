package amenitycatalog

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Delete(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool) {
	id, err := typeID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := set.Amenities.DeleteType(scope(ownerID, admin), id); errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		slog.Error("deleting amenity type", "owner_id", ownerID, "amenity_type_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	flashAndRedirect(w, r, ownerID, admin, helpers.FlashSuccess, "Amenidad y valores asignados eliminados.")
}
