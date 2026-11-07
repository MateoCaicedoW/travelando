package reservations

import (
	"net/http"

	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
)

type reservationFormState = helpers.ReservationFormState

func reservationFromForm(r *http.Request) (models.ReservationCreate, validate.Errors) {
	return helpers.ReservationFromForm(r, services.MaxAmountCOP)
}
