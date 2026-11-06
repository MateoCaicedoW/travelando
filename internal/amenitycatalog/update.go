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

func Update(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := typeID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	amenityType, validationErrors := amenityTypeFromForm(r)
	amenityType.ID = id
	if validationErrors.HasAny() {
		renderManager(w, r, ownerID, admin, &components.AmenityManagerState{Type: amenityType, Errors: validationErrors}, http.StatusUnprocessableEntity)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	err = set.Amenities.UpdateType(scope(ownerID, admin), id, amenityType)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if errors.Is(err, services.ErrInvalidInput) {
		renderManager(w, r, ownerID, admin, &components.AmenityManagerState{Type: amenityType, Errors: validate.Errors{"name": {errors.New("Ya existe una amenidad con ese nombre.")}}}, http.StatusUnprocessableEntity)
		return
	}
	if err != nil {
		slog.Error("updating amenity type", "owner_id", ownerID, "amenity_type_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	flashAndRedirect(w, r, ownerID, admin, helpers.FlashSuccess, "Amenidad actualizada. Si cambiaste el tipo, sus valores anteriores se reiniciaron.")
}
