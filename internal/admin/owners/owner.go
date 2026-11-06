package owners

import "travelando/internal/system/models"

func ownerExists(owners []models.Owner, id int64) bool {
	for _, owner := range owners {
		if owner.ID == id {
			return true
		}
	}
	return false
}

func ownerByID(owners []models.Owner, id int64) (models.Owner, bool) {
	for _, owner := range owners {
		if owner.ID == id {
			return owner, true
		}
	}
	return models.Owner{}, false
}

func ownerConflictMessage(state *ownerFormState) {
	addOwnerFormError(state, "cedula", "La cédula o celular ya pertenece a otro dueño.")
}
