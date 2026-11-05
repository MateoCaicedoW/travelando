package calendar

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Calendar
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	now := time.Now()
	year, err := calendarParam(r, "year", now.Year())
	if err != nil {
		http.Error(w, "Año inválido", http.StatusBadRequest)
		return
	}
	monthNumber, err := calendarParam(r, "month", int(now.Month()))
	if err != nil || monthNumber < 1 || monthNumber > 12 {
		http.Error(w, "Mes inválido", http.StatusBadRequest)
		return
	}
	month := time.Month(monthNumber)
	data, err := svc.Calendar(year, month)
	if err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			http.Error(w, "Fecha inválida", http.StatusBadRequest)
			return
		}
		slog.Error("loading admin calendar", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	view := newCalendarView(year, month, data, now)
	if helpers.IsHTMX(r) {
		if err := calendarViewEl(view).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Calendario | Travelando", Portal: system.PortalAdmin, Yield: calendarPageEl(view)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func calendarParam(r *http.Request, key string, fallback int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
