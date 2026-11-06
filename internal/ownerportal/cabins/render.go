package cabins

import (
	"log/slog"
	"net/http"
	"travelando/internal/ownerportal/auth"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func ownerCabinSuccess(w http.ResponseWriter, r *http.Request, page int, message string) {
	if hxhttp.IsRequest(r.Header) {
		renderOwnerCabins(w, r, page, http.StatusOK, helpers.FlashSuccess, message, true)
		return
	}
	if err := helpers.SetFlash(r, helpers.FlashSuccess, message); err != nil {
		http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
		return
	}
	helpers.Redirect(w, r, ownerCabinsURL(page))
}

func renderOwnerCabins(w http.ResponseWriter, r *http.Request, page, status int, flashLevel, flashMessage string, closeModal bool) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Owners == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	result, err := set.Owners.OwnerCabinsPage(auth.OwnerIDFromSession(r), page, ownerCabinsPerPage)
	if err != nil {
		slog.Error("loading owner cabins", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	content := ownerCabinsContentEl(result)
	if closeModal {
		content = Group{content, Div(ID("modal"), hx.SwapOOB("delete"))}
	}
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
	pageEl := system.Layout{Title: "Mis cabañas | Travelando", Portal: system.PortalOwner, Yield: content}.New(r)
	if err := pageEl.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func ownerCabinModalResponse(w http.ResponseWriter, r *http.Request, cabin models.Cabin, state *ownerCabinFormState, ownerCabinErr validate.Errors, editing bool, page, status int) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	amenityTypes, err := set.Amenities.TypesForOwner(auth.OwnerIDFromSession(r))
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if state == nil {
		state = &ownerCabinFormState{}
	}
	state.AmenityTypes = amenityTypes
	modal := ownerCabinDialogEl(cabin, ownerCabinErr, cabinValues(state), state.AmenityTypes, editing, page, false)
	if hxhttp.IsRequest(r.Header) {
		w.WriteHeader(status)
		if err := modal.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	pageEl := system.Layout{Title: "Cabaña | Travelando", Portal: system.PortalOwner, Yield: modal}.New(r)
	w.WriteHeader(status)
	if err := pageEl.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func ownerCabinFormError(w http.ResponseWriter, r *http.Request, page int, cabin models.Cabin, state *ownerCabinFormState, editing bool) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Owners == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	result, err := set.Owners.OwnerCabinsPage(auth.OwnerIDFromSession(r), page, ownerCabinsPerPage)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	content := ownerCabinsContentEl(result)
	if hxhttp.IsRequest(r.Header) {
		content = Group{content, ownerCabinDialogEl(cabin, state.Errors, state.Values, state.AmenityTypes, editing, page, true)}
		w.WriteHeader(http.StatusUnprocessableEntity)
		if err := content.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusUnprocessableEntity)
	pageEl := system.Layout{Title: "Cabaña | Travelando", Portal: system.PortalOwner, Yield: Group{content, ownerCabinDialogEl(cabin, state.Errors, state.Values, state.AmenityTypes, editing, page, false)}}.New(r)
	if err := pageEl.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
