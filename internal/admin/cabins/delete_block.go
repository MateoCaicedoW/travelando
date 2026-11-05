package cabins

import (
	"net/http"
	"strconv"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"
)

func DeleteBlock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	blockID, err := strconv.ParseInt(r.PathValue("blockID"), 10, 64)
	if err != nil || blockID < 1 {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Blocks
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := svc.DeleteBlock(id, blockID); err != nil {
		http.NotFound(w, r)
		return
	}
	helpers.Redirect(w, r, "/admin/cabins/"+strconv.FormatInt(id, 10)+"/edit")
}
