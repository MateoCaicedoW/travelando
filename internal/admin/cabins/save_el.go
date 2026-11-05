package cabins

import (
	"travelando/internal/system/models"

	"maragu.dev/gomponents"
)

func saveErrorEl(cabin models.Cabin, blocks []models.CabinBlock, bookings []models.Inquiry, owners []models.Owner, state *cabinFormState) gomponents.Node {
	if cabin.ID > 0 {
		return editEl(cabin, blocks, bookings, owners, state)
	}
	return newEl(cabin, owners, state)
}
