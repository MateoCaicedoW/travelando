package cabins

import (
	"net/url"

	admincabins "travelando/internal/admin/cabins"
	"travelando/internal/system/components"
	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
	. "maragu.dev/gomponents"
)

func ownerCabinDialogEl(cabin models.Cabin, fieldErrors validate.Errors, values url.Values, amenityTypes []models.AmenityType, editing bool, page int, oob bool) Node {
	label := "Agregar cabaña"
	if editing {
		label = "Editar cabaña"
	}
	swapOOB := ""
	if oob {
		swapOOB = "outerHTML"
	}
	return components.CustomDialog(components.DialogProps{Label: label, SwapOOB: swapOOB, Yield: admincabins.OwnerCabinFormEl(cabin, fieldErrors, values, amenityTypes, editing, page)})
}
