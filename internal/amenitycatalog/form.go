package amenitycatalog

import (
	"errors"
	"net/http"
	"strconv"

	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

func amenityTypeFromForm(r *http.Request) (models.AmenityType, validate.Errors) {
	errors := form.Validate(r, validate.Fields(
		validate.Field("name", validate.Required("Ingresa el nombre."), validate.MaxLength(80, "Usa máximo 80 caracteres.")),
		validate.Field("value_type", validValueType),
	))
	return models.AmenityType{Name: r.FormValue("name"), ValueType: models.AmenityValueType(r.FormValue("value_type"))}, errors
}

func validValueType(values []string) error {
	if len(values) == 1 && (values[0] == string(models.AmenityBoolean) || values[0] == string(models.AmenityQuantity)) {
		return nil
	}
	return errors.New("Selecciona sí/no o cantidad.")
}

func typeID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, services.ErrNotFound
	}
	return id, nil
}
