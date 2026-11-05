package cabins

import (
	"net/url"

	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
	. "maragu.dev/gomponents"
)

func OwnerCabinFormEl(cabin models.Cabin, fieldErrors validate.Errors, values url.Values, editing bool, page int) Node {
	state := &cabinFormState{Errors: fieldErrors, Values: values}
	return cabinSharedFormEl(cabin, nil, nil, nil, state, editing, true, page)
}
