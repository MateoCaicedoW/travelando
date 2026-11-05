package cabins

import (
	"log/slog"
	"net/http"

	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func New(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owners, err := svc.Owners()
	if err != nil {
		slog.Error("loading cabin owners", "error", err)
		http.Error(w, "No se pudo cargar el formulario", http.StatusInternalServerError)
		return
	}
	dialog := components.CustomDialog(components.DialogProps{Yield: newEl(defaultNewCabin(), owners, nil)})
	if err := dialog.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func defaultNewCabin() models.Cabin {
	return models.Cabin{Active: true, IncludedGuests: 2, AcceptsDayUse: true, AcceptsOvernight: true, DayStart: "09:00", DayEnd: "18:00", CheckIn: "15:00", CheckOut: "12:00", CleaningHours: 3}
}

func DefaultNewCabin() models.Cabin { return defaultNewCabin() }
