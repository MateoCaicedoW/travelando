package home

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/system"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Cabin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Cabins
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	cabin, err := svc.Cabin(id)
	if errors.Is(err, services.ErrNotFound) || err == nil && (cabin == nil || !cabin.Active || !cabin.OwnerActive || cabin.OwnerID < 1 || cabin.OwnerWhatsApp == "") {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		slog.Error("loading cabin page", "error", err)
		http.Error(w, "No se pudo cargar la cabaña", http.StatusInternalServerError)
		return
	}
	data := CabinQueryData(r.URL.Query())
	page := system.Layout{
		Title:       cabin.Name + " | Travelando",
		Description: cabin.Description,
		Yield:       CabinEl(*cabin, data),
	}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
