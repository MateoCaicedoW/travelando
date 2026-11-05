package ownerportal

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"

	lucide "github.com/eduardolat/gomponents-lucide"
	admincabins "travelando/internal/admin/cabins"
	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/helpers"
	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form/validate"
	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

const ownerCabinsPerPage = 10

type ownerCabinFormState struct {
	Errors       validate.Errors
	Values       url.Values
	AmenityTypes []models.AmenityType
}

func Cabins(w http.ResponseWriter, r *http.Request) {
	page := ownerCabinPageParam(r)
	renderOwnerCabins(w, r, page, http.StatusOK, "", "", false)
}

func NewCabin(w http.ResponseWriter, r *http.Request) {
	page := ownerCabinPageParam(r)
	cabin := admincabins.DefaultNewCabin()
	ownerCabinModalResponse(w, r, cabin, nil, nil, false, page, http.StatusOK)
}

func EditCabin(w http.ResponseWriter, r *http.Request) {
	id, err := ownerCabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	if set.Cabins == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	cabin, err := set.Cabins.Cabin(id)
	if errors.Is(err, services.ErrNotFound) || err == nil && cabin.OwnerID != ownerIDFromSession(r) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	ownerCabinModalResponse(w, r, *cabin, nil, nil, true, ownerCabinPageParam(r), http.StatusOK)
}

func CreateCabin(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	ownerID := ownerIDFromSession(r)
	page := ownerCabinPageParam(r)
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil || set.Cabins == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	amenityTypes, err := set.Amenities.TypesForOwner(ownerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabin, fieldErrors := admincabins.OwnerCabinFromForm(r, 0, ownerID)
	cabinAmenities, amenityErrors := admincabins.CabinAmenitiesFromForm(r, amenityTypes)
	cabin.Amenities = cabinAmenities
	fieldErrors = admincabins.MergeCabinFormErrors(fieldErrors, amenityErrors)
	state := &ownerCabinFormState{Errors: fieldErrors, Values: r.Form, AmenityTypes: amenityTypes}
	if fieldErrors.HasAny() {
		ownerCabinFormError(w, r, page, cabin, state, false)
		return
	}
	cabin.Active = true
	if err := set.Cabins.CreateCabinForOwner(ownerID, &cabin); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			addOwnerCabinFormError(state, "name", "Revisa los datos de la cabaña.")
			ownerCabinFormError(w, r, page, cabin, state, false)
			return
		}
		slog.Error("owner creating cabin", "owner_id", ownerID, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	ownerCabinSuccess(w, r, page, "Cabaña agregada y publicada.")
}

func UpdateCabin(w http.ResponseWriter, r *http.Request) {
	if err := helpers.ParseForm(w, r); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	id, err := ownerCabinIDFromPath(r)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	ownerID := ownerIDFromSession(r)
	page := ownerCabinPageParam(r)
	set := services.ServiceSetFromCtx(r.Context())
	if set.Amenities == nil || set.Cabins == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	amenityTypes, err := set.Amenities.TypesForOwner(ownerID)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	cabin, fieldErrors := admincabins.OwnerCabinFromForm(r, id, ownerID)
	cabinAmenities, amenityErrors := admincabins.CabinAmenitiesFromForm(r, amenityTypes)
	cabin.Amenities = cabinAmenities
	fieldErrors = admincabins.MergeCabinFormErrors(fieldErrors, amenityErrors)
	cabin.Active = r.FormValue("active") == "on"
	state := &ownerCabinFormState{Errors: fieldErrors, Values: r.Form, AmenityTypes: amenityTypes}
	if fieldErrors.HasAny() {
		ownerCabinFormError(w, r, page, cabin, state, true)
		return
	}
	if err := set.Cabins.UpdateCabinForOwner(ownerID, &cabin); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if errors.Is(err, services.ErrInvalidInput) {
			addOwnerCabinFormError(state, "name", "Revisa los datos de la cabaña.")
			ownerCabinFormError(w, r, page, cabin, state, true)
			return
		}
		slog.Error("owner updating cabin", "owner_id", ownerID, "cabin_id", id, "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	ownerCabinSuccess(w, r, page, "Cabaña actualizada.")
}

func ownerCabinSuccess(w http.ResponseWriter, r *http.Request, page int, message string) {
	if helpers.IsHTMX(r) {
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
	result, err := set.Owners.OwnerCabinsPage(ownerIDFromSession(r), page, ownerCabinsPerPage)
	if err != nil {
		slog.Error("loading owner cabins", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	content := ownerCabinsContentEl(result)
	if closeModal {
		content = Group{content, Div(ID("modal"), Attr("hx-swap-oob", "delete"))}
	}
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
	amenityTypes, err := set.Amenities.TypesForOwner(ownerIDFromSession(r))
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if state == nil {
		state = &ownerCabinFormState{}
	}
	state.AmenityTypes = amenityTypes
	modal := ownerCabinDialogEl(cabin, ownerCabinErr, cabinValues(state), state.AmenityTypes, editing, page, false)
	if helpers.IsHTMX(r) {
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
	result, err := set.Owners.OwnerCabinsPage(ownerIDFromSession(r), page, ownerCabinsPerPage)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	content := ownerCabinsContentEl(result)
	if helpers.IsHTMX(r) {
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

func addOwnerCabinFormError(state *ownerCabinFormState, field, message string) {
	if state.Errors == nil {
		state.Errors = validate.Errors{}
	}
	state.Errors[field] = append(state.Errors[field], errors.New(message))
}

func cabinValues(state *ownerCabinFormState) url.Values {
	if state == nil {
		return nil
	}
	return state.Values
}

func ownerCabinDialogEl(cabin models.Cabin, fieldErrors validate.Errors, values url.Values, amenityTypes []models.AmenityType, editing bool, page int, oob bool) Node {
	label := "Agregar cabaña"
	if editing {
		label = "Editar cabaña"
	}
	swapOOB := ""
	if oob {
		swapOOB = "outerHTML"
	}
	return components.CustomDialog(components.DialogProps{Label: label, SwapOOB: swapOOB, Yield: admincabins.OwnerCabinFormEl(cabin, fieldErrors, values, amenityTypes, editing, page)})
}

func ownerCabinsContentEl(result services.CabinPageResult) Node {
	return Div(ID("owner-content"), Class("mx-auto w-full max-w-7xl px-5 py-8 sm:px-8 sm:py-12"),
		Div(Class("flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between"),
			components.PageHeading("Mis cabañas", "Administra los datos y publicación de tus cabañas."),
			Button(Type("button"), Attr("hx-get", ownerNewCabinURL(result.Page)), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Attr("hx-disabled-elt", "this"), Class("inline-flex min-h-10 items-center rounded-lg bg-brand-dark px-4 py-2 text-sm font-extrabold text-white hover:brightness-90"), Text("Agregar cabaña")),
		),
		If(len(result.Cabins) == 0, Div(Class("mt-6 rounded-2xl border border-line bg-white px-5 py-12 text-center text-sm text-muted"), Text("Todavía no tienes cabañas. Agrega una para publicarla."))),
		If(len(result.Cabins) > 0, Div(ID("owner-cabin-cards"), Class("mt-6 grid gap-5 sm:grid-cols-2 xl:grid-cols-3"),
			Map(result.Cabins, func(cabin models.Cabin) Node { return ownerCabinCardEl(cabin, result.Page) }),
		)),
		If(result.TotalPages > 1, Nav(Attr("aria-label", "Paginación de mis cabañas"), Class("mt-8 grid grid-cols-[minmax(0,1fr)_auto_minmax(0,1fr)] items-center gap-2 border-t border-line pt-5 sm:gap-4"),
			If(result.Page > 1, A(Href(ownerCabinsURL(result.Page-1)), Attr("hx-get", ownerCabinsURL(result.Page-1)), Attr("hx-target", "#owner-content"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("inline-flex min-h-10 items-center justify-center rounded-lg border border-line px-3 text-sm font-bold text-ink hover:border-brand"), Text("← Anterior"))),
			P(Class("justify-self-center rounded-full bg-cream px-3 py-2 text-center text-xs font-extrabold text-brand-dark"), Textf("Página %d de %d", result.Page, result.TotalPages)),
			If(result.Page < result.TotalPages, A(Href(ownerCabinsURL(result.Page+1)), Attr("hx-get", ownerCabinsURL(result.Page+1)), Attr("hx-target", "#owner-content"), Attr("hx-swap", "outerHTML"), Attr("hx-push-url", "true"), Class("col-start-3 inline-flex min-h-10 items-center justify-center rounded-lg border border-line px-3 text-sm font-bold text-ink hover:border-brand"), Text("Siguiente →"))),
		)),
	)
}

func ownerCabinCardEl(cabin models.Cabin, page int) Node {
	status, statusStyle := "Publicada", "bg-forest-soft text-forest"
	statusDot := "bg-forest"
	if !cabin.Active {
		status, statusStyle = "Inactiva", "bg-cream text-muted"
		statusDot = "bg-muted"
	}
	image := Div(Class("grid aspect-[16/10] place-items-center bg-cream text-muted"),
		lucide.Image(Class("size-8"), Attr("aria-hidden", "true")),
		Span(Class("sr-only"), Text("Sin foto")),
	)
	if cabin.CoverImage != "" {
		image = Img(Src(cabin.CoverImage), Alt("Foto de "+cabin.Name), Attr("loading", "lazy"), Class("aspect-[16/10] w-full object-cover"))
	}
	return Article(Class("overflow-hidden rounded-2xl border border-line bg-white transition hover:border-brand/60"),
		Div(Class("relative overflow-hidden bg-cream"),
			image,
			Span(Class("absolute right-3 top-3 inline-flex items-center gap-2 rounded-full bg-white px-3 py-1.5 text-xs font-bold shadow-sm "+statusStyle),
				Span(Attr("aria-hidden", "true"), Class("size-2 rounded-full "+statusDot)),
				Text(status),
			),
		),
		Div(Class("p-5"),
			Div(Class("flex items-start gap-3"),
				Div(Class("min-w-0 flex-1"),
					H2(Class("truncate text-xl font-extrabold text-ink"), Text(cabin.Name)),
					P(Class("mt-1 flex items-center gap-1.5 text-sm font-semibold text-muted"),
						lucide.MapPin(Class("size-4 shrink-0 text-brand-dark"), Attr("aria-hidden", "true")), Text(cabin.Location)),
				),
			),
			If(cabin.Description != "", P(Class("mt-3 line-clamp-2 min-h-12 text-sm leading-6 text-muted"), Text(cabin.Description))),
			Div(Class("mt-4 grid grid-cols-2 gap-3 border-y border-line py-3"),
				Div(P(Class("text-xs font-semibold text-muted"), Text("Pasadía")), P(Class("mt-1 text-lg font-extrabold text-ink"), Text("$"+helpers.FormatCOP(cabin.DayPrice)))),
				Div(P(Class("text-xs font-semibold text-muted"), Text("Por noche")), P(Class("mt-1 text-lg font-extrabold text-ink"), Text("$"+helpers.FormatCOP(cabin.NightPrice)))),
			),
			Div(Class("flex flex-wrap items-center justify-between gap-2 py-3 text-xs font-semibold text-muted"),
				Span(Class("inline-flex items-center gap-1.5"), lucide.Users(Class("size-4"), Attr("aria-hidden", "true")), Textf("%d huéspedes incluidos", cabin.IncludedGuests)),
				If(cabin.ExtraGuestPrice > 0, Span(Text("+$"+helpers.FormatCOP(cabin.ExtraGuestPrice)+" por adicional"))),
			),
			Button(Type("button"), Attr("hx-get", ownerEditCabinURL(cabin.ID, page)), Attr("hx-target", "body"), Attr("hx-swap", "beforeend"), Class("inline-flex min-h-10 w-full items-center justify-center gap-2 rounded-lg border border-line px-3 py-2 text-sm font-bold text-brand-dark transition hover:border-brand hover:bg-cream focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand-dark"),
				lucide.Pencil(Class("size-4"), Attr("aria-hidden", "true")), Text("Editar cabaña")),
		),
	)
}

func ownerCabinPageParam(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func ownerCabinsURL(page int) string   { return "/owner/cabins?page=" + strconv.Itoa(page) }
func ownerNewCabinURL(page int) string { return "/owner/cabins/new?page=" + strconv.Itoa(page) }
func ownerEditCabinURL(id int64, page int) string {
	return "/owner/cabins/" + strconv.FormatInt(id, 10) + "/edit?page=" + strconv.Itoa(page)
}

func ownerCabinIDFromPath(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, services.ErrNotFound
	}
	return id, nil
}
