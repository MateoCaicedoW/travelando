package reservations

import (
	reservationui "travelando/internal/reservations"
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
)

func reservationDetailsEl(inquiry models.Inquiry) Node {
	return reservationui.ReservationDetailsEl(inquiry)
}
