package owners

import (
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"
	. "maragu.dev/gomponents/html"
)

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
		content = Group{content, Div(ID("global-flash-container"), hx.SwapOOB("true"), components.FlashToast(flashMessage, flashLevel))}
	}
	w.WriteHeader(status)
	if hxhttp.IsRequest(r.Header) {
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
	if hxhttp.IsRequest(r.Header) {
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
	if !hxhttp.IsRequest(r.Header) {
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
		Div(ID("modal"), hx.SwapOOB("delete")),
		Div(ID("global-flash-container"), hx.SwapOOB("true"), components.FlashToast("Dueño creado.", helpers.FlashSuccess)),
	}
	if err := response.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
