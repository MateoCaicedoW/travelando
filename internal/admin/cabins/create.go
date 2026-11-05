package cabins

import (
	"net/http"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

func Create(w http.ResponseWriter, r *http.Request) {
	saveCabin(w, r, createCabin)
}

func createCabin(svc services.CabinService, cabin *models.Cabin) error {
	return svc.CreateCabin(cabin)
}
