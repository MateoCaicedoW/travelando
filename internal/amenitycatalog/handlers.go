package amenitycatalog

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
)

func Index(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool) {
	if !admin && ownerID < 1 {
		http.NotFound(w, r)
		return
	}
	renderManager(w, r, ownerID, admin, nil, http.StatusOK)
}

func renderManager(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool, state *components.AmenityManagerState, status int) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	types, err := set.Amenities.TypesForOwner(ownerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	types = managedTypes(types, ownerID, admin)
	page, title, description := pageOptions(admin)
	layout := system.Layout{Title: title + " | Travelando", Portal: portal(admin), Yield: components.AmenityManager(title, description, page, types, state)}.New(r)
	w.WriteHeader(status)
	if err := layout.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

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

func Delete(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool) {
	id, err := typeID(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	if err := set.Amenities.DeleteType(scope(ownerID, admin), id); errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		slog.Error("deleting amenity type", "owner_id", ownerID, "amenity_type_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	flashAndRedirect(w, r, ownerID, admin, helpers.FlashSuccess, "Amenidad y valores asignados eliminados.")
}

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

func managedTypes(types []models.AmenityType, ownerID int64, admin bool) []models.AmenityType {
	managed := make([]models.AmenityType, 0, len(types))
	for _, amenityType := range types {
		if admin && amenityType.OwnerID == 0 || !admin && amenityType.OwnerID == ownerID {
			managed = append(managed, amenityType)
		}
	}
	return managed
}

func scope(ownerID int64, admin bool) int64 {
	if admin {
		return 0
	}
	return ownerID
}

func portal(admin bool) string {
	if admin {
		return system.PortalAdmin
	}
	return system.PortalOwner
}

func pageOptions(admin bool) (string, string, string) {
	if admin {
		return "/admin/amenities", "Amenidades predeterminadas", "Catálogo global visible y reutilizable por todos los dueños."
	}
	return "/owner/amenities", "Mis amenidades", "Tus tipos personalizados se reutilizan en todas tus cabañas. Las amenidades predeterminadas también están disponibles al editar una cabaña."
}

func flashAndRedirect(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool, level, message string) {
	if err := helpers.SetFlash(r, level, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	path, _, _ := pageOptions(admin)
	helpers.Redirect(w, r, path)
}

func typeID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, services.ErrNotFound
	}
	return id, nil
}
