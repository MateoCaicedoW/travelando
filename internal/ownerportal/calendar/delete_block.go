package calendar

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func DeleteBlock(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	blockID, err := strconv.ParseInt(r.PathValue("blockID"), 10, 64)
	if err != nil || blockID < 1 {
		http.NotFound(w, r)
		return
	}
	cabinID, err := strconv.ParseInt(r.FormValue("cabin_id"), 10, 64)
	if err != nil || cabinID < 1 {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Blocks
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := svc.DeleteExternalBlock(auth.OwnerIDFromSession(r), cabinID, blockID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("deleting owner external block", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if hxhttp.IsRequest(r.Header) {
		if err := renderOwnerCalendar(w, r, http.StatusOK, helpers.FlashSuccess, "Reserva externa quitada del calendario.", false); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, "Reserva externa quitada del calendario."); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, ownerMonthURLForDate(r.FormValue("date_from")))
}
