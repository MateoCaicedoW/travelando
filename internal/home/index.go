package home

import (
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	data := searchPageData(r, services.ServiceSetFromCtx(r.Context()).Availability)
	if data.Error != "" {
		slog.Error("searching available cabins", "error", data.Error)
		if err := helpers.SetFlash(r, helpers.FlashError, "No se pudo cargar la disponibilidad. Intenta de nuevo."); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
	}
	if helpers.IsHTMX(r) && r.Header.Get("HX-History-Restore-Request") != "true" {
		if err := IndexEl(data).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Travelando | Tu próximo plan está más cerca", Description: "Encuentra cabañas disponibles para tu próxima escapada.", Yield: IndexEl(data)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
