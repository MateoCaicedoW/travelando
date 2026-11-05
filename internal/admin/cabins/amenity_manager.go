package cabins

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
)

func CreateCabinAmenityTypeHandler(w http.ResponseWriter, r *http.Request) {
	createCabinAmenityType(w, r, 0, false)
}

func CreateCabinAmenityType(w http.ResponseWriter, r *http.Request, ownerID int64, ownerMode bool) {
	createCabinAmenityType(w, r, ownerID, ownerMode)
}

func createCabinAmenityType(w http.ResponseWriter, r *http.Request, ownerID int64, ownerMode bool) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	amenityType, fieldErrors := cabinAmenityTypeFromForm(r)
	state := &components.AmenityManagerState{Type: amenityType, Errors: fieldErrors}
	if fieldErrors.HasAny() {
		respondCabinAmenityFields(w, r, ownerID, ownerMode, state)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	scope := cabinAmenityScope(ownerID, ownerMode)
	if err := set.Amenities.CreateType(scope, &amenityType); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			state.Errors = validate.Errors{"new_amenity_name": {errors.New("Ya existe una amenidad con ese nombre.")}}
			respondCabinAmenityFields(w, r, ownerID, ownerMode, state)
			return
		}
		slog.Error("creating amenity from cabin form", "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	state.Type = models.AmenityType{ValueType: models.AmenityBoolean}
	state.Message = "Amenidad agregada."
	respondCabinAmenityFields(w, r, ownerID, ownerMode, state)
}

func cabinAmenityTypeFromForm(r *http.Request) (models.AmenityType, validate.Errors) {
	errors := form.Validate(r, validate.Fields(
		validate.Field("new_amenity_name", validate.Required("Ingresa el nombre."), validate.MaxLength(80, "Usa máximo 80 caracteres.")),
		validate.Field("new_amenity_value_type", validCabinAmenityType),
	))
	return models.AmenityType{
		Name:      r.FormValue("new_amenity_name"),
		ValueType: models.AmenityValueType(r.FormValue("new_amenity_value_type")),
	}, errors
}

func validCabinAmenityType(values []string) error {
	if len(values) == 1 && (values[0] == string(models.AmenityBoolean) || values[0] == string(models.AmenityQuantity)) {
		return nil
	}
	return errors.New("Selecciona sí/no o cantidad.")
}

func cabinAmenityScope(ownerID int64, ownerMode bool) int64 {
	if ownerMode {
		return ownerID
	}
	return 0
}

func respondCabinAmenityFields(w http.ResponseWriter, r *http.Request, ownerID int64, ownerMode bool, state *components.AmenityManagerState) {
	if !helpers.IsHTMX(r) {
		level := helpers.FlashSuccess
		if state.Errors.HasAny() {
			level = helpers.FlashError
			state.Message = "Revisa el nombre y el tipo de amenidad."
		}
		if err := helpers.SetFlash(r, level, state.Message); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		path := "/admin/amenities"
		if ownerMode {
			path = "/owner/amenities"
		}
		helpers.Redirect(w, r, path)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	cabinOwnerID := ownerID
	if !ownerMode {
		cabinOwnerID, _ = strconv.ParseInt(r.FormValue("owner_id"), 10, 64)
	}
	types, err := set.Amenities.TypesForOwner(cabinOwnerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	fields := amenityFieldsEl(models.Cabin{OwnerID: cabinOwnerID}, types, r.Form, nil, ownerMode, false, state)
	w.WriteHeader(http.StatusOK)
	if err := fields.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
