package cabins

import (
	"errors"
	"net/url"

	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
)

const ownerCabinsPerPage = 10

type ownerCabinFormState struct {
	Errors       validate.Errors
	Values       url.Values
	AmenityTypes []models.AmenityType
}

func addOwnerCabinFormError(state *ownerCabinFormState, field, message string) {
	if state.Errors == nil {
		state.Errors = validate.Errors{}
	}
	state.Errors[field] = append(state.Errors[field], errors.New(message))
}

func cabinValues(state *ownerCabinFormState) url.Values {
	if state == nil {
		return nil
	}
	return state.Values
}
