package cabins

import (
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
)

func editModalEl(cabin models.Cabin, blocks []models.CabinBlock, bookings []models.Inquiry, owners []models.Owner, amenityTypes []models.AmenityType, state *cabinFormState) Node {
	return cabinSharedFormEl(cabin, blocks, bookings, owners, amenityTypes, state, true, false, 0)
}
