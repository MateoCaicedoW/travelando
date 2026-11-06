package cabins

import (
	"net/http"
	"strconv"
	"strings"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func DeletePhoto(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	photoURL := r.FormValue("photo_url")
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil || set.Storage == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := set.Cabins.DeleteCabinPhoto(id, photoURL); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	_ = set.Storage.DeletePhoto(strings.TrimPrefix(photoURL, "/uploads/"))
	w.Header().Set("HX-Trigger", "photo-deleted")
	w.WriteHeader(http.StatusOK)
}
