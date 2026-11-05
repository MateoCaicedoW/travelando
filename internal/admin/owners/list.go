package owners

import (
	"errors"
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owners, err := svc.Owners()
	if err != nil {
		slog.Error("loading owners", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderOwnersContent(w, r, owners, 0, nil, http.StatusOK, "", "")
}

func New(w http.ResponseWriter, r *http.Request) {
	dialog := ownerCreateDialogEl(nil, false)
	if helpers.IsHTMX(r) {
		if err := dialog.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Agregar dueño | Travelando", Portal: system.PortalAdmin, Yield: dialog}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func Edit(w http.ResponseWriter, r *http.Request) {
	id, err := ownerIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	owners, err := svc.Owners()
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if !ownerExists(owners, id) {
		http.NotFound(w, r)
		return
	}
	renderOwnersContent(w, r, owners, id, nil, http.StatusOK, "", "")
}

func renderOwnersPage(w http.ResponseWriter, r *http.Request, svc services.OwnerService, editID int64, editState *ownerFormState, status int, flashLevel, flashMessage string) {
	owners, err := svc.Owners()
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderOwnersContent(w, r, owners, editID, editState, status, flashLevel, flashMessage)
}

func renderOwnersContent(w http.ResponseWriter, r *http.Request, owners []models.Owner, editID int64, editState *ownerFormState, status int, flashLevel, flashMessage string) {
	content := ownersEl(owners, editID, editState)
	if flashMessage != "" {
		content = Group{content, Div(ID("global-flash-container"), Attr("hx-swap-oob", "true"), components.FlashToast(flashMessage, flashLevel))}
	}
	w.WriteHeader(status)
	if helpers.IsHTMX(r) {
		if err := content.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Dueños | Travelando", Portal: system.PortalAdmin, Yield: content}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func renderOwnerCreateError(w http.ResponseWriter, r *http.Request, svc services.OwnerService, state *ownerFormState, status int) {
	owners, err := svc.Owners()
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	content := ownersEl(owners, 0, nil)
	if helpers.IsHTMX(r) {
		content = Group{content, ownerCreateDialogEl(state, true)}
		w.WriteHeader(status)
		if err := content.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(status)
	page := system.Layout{Title: "Agregar dueño | Travelando", Portal: system.PortalAdmin, Yield: Group{content, ownerCreateDialogEl(state, false)}}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func renderOwnerCreateSuccess(w http.ResponseWriter, r *http.Request, svc services.OwnerService) {
	if !helpers.IsHTMX(r) {
		if err := helpers.SetFlash(r, helpers.FlashSuccess, "Dueño creado."); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
		helpers.Redirect(w, r, "/admin/owners")
		return
	}
	owners, err := svc.Owners()
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	response := Group{
		ownersEl(owners, 0, nil),
		Div(ID("modal"), Attr("hx-swap-oob", "delete")),
		Div(ID("global-flash-container"), Attr("hx-swap-oob", "true"), components.FlashToast("Dueño creado.", helpers.FlashSuccess)),
	}
	if err := response.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func SetActive(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := ownerIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	activeValue := r.FormValue("active")
	if activeValue != "0" && activeValue != "1" {
		http.Error(w, "Estado inválido", http.StatusBadRequest)
		return
	}
	svc := services.ServiceSetFromCtx(r.Context()).Owners
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	active := activeValue == "1"
	if err := svc.SetOwnerActive(id, active); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		slog.Error("updating owner active state", "owner_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	message := "Dueño activado."
	if !active {
		message = "Dueño desactivado; cabañas se mantienen guardadas y dejan de mostrarse."
	}
	if helpers.IsHTMX(r) {
		renderOwnersPage(w, r, svc, 0, nil, http.StatusOK, helpers.FlashSuccess, message)
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, "/admin/owners")
}

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
