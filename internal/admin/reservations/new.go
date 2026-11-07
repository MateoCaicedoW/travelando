package reservations

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func New(w http.ResponseWriter, r *http.Request) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	result, err := set.Cabins.CabinsPage(1, 999)
	if err != nil {
		slog.Error("loading cabins for new reservation", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabins := activeCabins(result.Cabins)
	state := &reservationFormState{Cabins: cabins}
	dialog := components.CustomDialog(components.DialogProps{
		Label: "Nueva reserva",
		Yield: newReservationFormEl(state),
	})
	if err := dialog.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func Create(w http.ResponseWriter, r *http.Request) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil || set.Inquiries == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		server.Error(w, err, http.StatusBadRequest)
		return
	}
	data, fieldErrors := reservationFromForm(r)
	result, err := set.Cabins.CabinsPage(1, 999)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabins := activeCabins(result.Cabins)

	if fieldErrors.HasAny() {
		state := &reservationFormState{Errors: fieldErrors, Values: r.Form, Cabins: cabins}
		dialog := components.CustomDialog(components.DialogProps{
			Label: "Nueva reserva",
			Yield: newReservationFormEl(state),
		})

		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := dialog.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}

		return
	}

	if _, err := set.Inquiries.CreateReservation(data.CabinID, &data); err != nil {
		if errors.Is(err, services.ErrUnavailable) {
			if fieldErrors == nil {
				fieldErrors = validate.Errors{}
			}
			fieldErrors["date_from"] = append(fieldErrors["date_from"], errors.New("Las fechas seleccionadas ya están ocupadas para esta cabaña."))
		}

		state := &reservationFormState{Errors: fieldErrors, Values: r.Form, Cabins: cabins}
		dialog := components.CustomDialog(components.DialogProps{
			Label: "Nueva reserva",
			Yield: newReservationFormEl(state),
		})
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := dialog.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}

	hxhttp.SetRefresh(w.Header())
}

func activeCabins(cabins []models.Cabin) []models.Cabin {
	out := make([]models.Cabin, 0, len(cabins))
	for _, c := range cabins {
		if c.Active && c.OwnerActive {
			out = append(out, c)
		}
	}
	return out
}
