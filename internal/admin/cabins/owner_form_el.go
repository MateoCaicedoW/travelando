package cabins

import (
	"net/url"

	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
	. "maragu.dev/gomponents"
)

func OwnerCabinFormEl(cabin models.Cabin, fieldErrors validate.Errors, values url.Values, amenityTypes []models.AmenityType, editing bool, page int) Node {
	state := &cabinFormState{Errors: fieldErrors, Values: values}
	state.AmenityTypes = amenityTypes
	return cabinSharedFormEl(cabin, nil, nil, nil, amenityTypes, state, editing, true, page)
}
