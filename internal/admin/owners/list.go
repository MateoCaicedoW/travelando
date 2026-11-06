package owners

import (
	"log/slog"
	"net/http"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owners, err := svc.Owners()
	if err != nil {
		slog.Error("loading owners", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderOwnersContent(w, r, owners, 0, nil, http.StatusOK, "", "")
}
