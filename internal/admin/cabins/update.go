package cabins

import (
	"net/http"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func Update(w http.ResponseWriter, r *http.Request) {
	saveCabin(w, r, updateCabin)
}

func updateCabin(svc services.CabinService, cabin *models.Cabin) error {
	return svc.UpdateCabin(cabin)
}
