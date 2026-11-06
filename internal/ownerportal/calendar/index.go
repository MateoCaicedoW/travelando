package calendar

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if err := renderOwnerCalendar(w, r, http.StatusOK, "", "", false); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			http.Error(w, "Mes o año inválido", http.StatusBadRequest)
			return
		}
		slog.Error("loading owner calendar", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
	}
}
