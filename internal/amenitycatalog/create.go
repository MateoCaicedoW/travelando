package amenitycatalog

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
)

func Create(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	amenityType, validationErrors := amenityTypeFromForm(r)
	if validationErrors.HasAny() {
		renderManager(w, r, ownerID, admin, &components.AmenityManagerState{Type: amenityType, Errors: validationErrors}, http.StatusUnprocessableEntity)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := set.Amenities.CreateType(scope(ownerID, admin), &amenityType); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			renderManager(w, r, ownerID, admin, &components.AmenityManagerState{Type: amenityType, Errors: validate.Errors{"name": {errors.New("Ya existe una amenidad con ese nombre.")}}}, http.StatusUnprocessableEntity)
			return
		}
		slog.Error("creating amenity type", "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	flashAndRedirect(w, r, ownerID, admin, helpers.FlashSuccess, "Amenidad agregada.")
}
