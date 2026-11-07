package reservations

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/ownerportal/auth"
	"travelando/internal/system/components"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func New(w http.ResponseWriter, r *http.Request) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Owners == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	ownerID := auth.OwnerIDFromSession(r)
	cabins, err := set.Owners.OwnerCabins(ownerID)
	if err != nil {
		slog.Error("loading owner cabins for new reservation", "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	state := &reservationFormState{Cabins: activeCabins(cabins)}
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
	if set.Owners == nil || set.Inquiries == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	ownerID := auth.OwnerIDFromSession(r)
	if err := r.ParseForm(); err != nil {
		server.Error(w, err, http.StatusBadRequest)
		return
	}
	data, fieldErrors := reservationFromForm(r)
	cabins, err := set.Owners.OwnerCabins(ownerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabins = activeCabins(cabins)

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

	if _, err := set.Inquiries.CreateReservationForOwner(ownerID, &data); err != nil {
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
		if c.Active {
			out = append(out, c)
		}
	}
	return out
}
