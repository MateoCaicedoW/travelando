package cabins

import (
	"errors"
	"net/http"
	"strconv"

	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
)

func Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil || set.Blocks == nil || set.Inquiries == nil || set.Owners == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	data, err := loadCabinEditData(set.Cabins, set.Blocks, set.Inquiries, set.Owners, id)
	if errors.Is(err, services.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if helpers.IsHTMX(r) {
		renderCabinEditDialog(w, data, nil, http.StatusOK)
		return
	}
	renderCabinEditPage(w, r, data, nil, http.StatusOK)
}

func renderCabinEditDialog(w http.ResponseWriter, data cabinEditData, state *cabinFormState, status int) {
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	dialog := components.CustomDialog(components.DialogProps{Label: "Editar cabaña", Yield: editModalEl(data.Cabin, data.Blocks, data.Bookings, data.Owners, state)})
	if err := dialog.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func refreshCabinEdit(w http.ResponseWriter, r *http.Request, id int64) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil || set.Blocks == nil || set.Inquiries == nil || set.Owners == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	data, err := loadCabinEditData(set.Cabins, set.Blocks, set.Inquiries, set.Owners, id)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	renderCabinEditPage(w, r, data, nil, http.StatusOK)
}

func renderCabinEditPage(w http.ResponseWriter, r *http.Request, data cabinEditData, state *cabinFormState, status int) {
	if status != http.StatusOK {
		w.WriteHeader(status)
	}
	if helpers.IsHTMX(r) {
		if err := editModalEl(data.Cabin, data.Blocks, data.Bookings, data.Owners, state).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Editar " + data.Cabin.Name + " | Travelando", Portal: system.PortalAdmin, Yield: editEl(data.Cabin, data.Blocks, data.Bookings, data.Owners, state)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
