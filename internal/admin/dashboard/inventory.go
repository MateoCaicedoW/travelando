package dashboard

import (
	"log/slog"
	"net/http"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Inventory(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Cabins
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	page, err := svc.CabinsPage(requestedCabinPage(r), cabinsPerPage)
	if err != nil {
		slog.Error("loading admin cabin page", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if err := inventoryEl(page).Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
