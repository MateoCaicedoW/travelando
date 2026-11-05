package cabins

import (
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
)

func editEl(cabin models.Cabin, blocks []models.CabinBlock, bookings []models.Inquiry, owners []models.Owner, state *cabinFormState) Node {
	return cabinSharedFormEl(cabin, blocks, bookings, owners, state, true, false, 0, false)
}

func editModalEl(cabin models.Cabin, blocks []models.CabinBlock, bookings []models.Inquiry, owners []models.Owner, state *cabinFormState) Node {
	return cabinSharedFormEl(cabin, blocks, bookings, owners, state, true, false, 0, true)
}
