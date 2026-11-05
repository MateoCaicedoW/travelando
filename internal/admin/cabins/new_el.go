package cabins

import (
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
)

func newEl(cabin models.Cabin, owners []models.Owner, amenityTypes []models.AmenityType, state *cabinFormState) Node {
	return cabinSharedFormEl(cabin, nil, nil, owners, amenityTypes, state, false, false, 0)
}
