package cabins

import (
	"travelando/internal/system/models"

	. "maragu.dev/gomponents"
)

func newEl(cabin models.Cabin, owners []models.Owner, state *cabinFormState) Node {
	return cabinSharedFormEl(cabin, nil, nil, owners, state, false, false, 0)
}
