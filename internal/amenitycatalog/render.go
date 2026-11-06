package amenitycatalog

import (
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

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

func pageOptions(admin bool) (string, string, string) {
	if admin {
		return "/admin/amenities", "Amenidades predeterminadas", "Catálogo global visible y reutilizable por todos los dueños."
	}
	return "/owner/amenities", "Mis amenidades", "Tus tipos personalizados se reutilizan en todas tus cabañas. Las amenidades predeterminadas también están disponibles al editar una cabaña."
}

func portal(admin bool) string {
	if admin {
		return system.PortalAdmin
	}
	return system.PortalOwner
}

func flashAndRedirect(w http.ResponseWriter, r *http.Request, ownerID int64, admin bool, level, message string) {
	if err := helpers.SetFlash(r, level, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	path, _, _ := pageOptions(admin)
	helpers.Redirect(w, r, path)
}
