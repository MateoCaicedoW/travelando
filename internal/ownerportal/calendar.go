package ownerportal

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"
	. "maragu.dev/gomponents/html"
	"travelando/internal/system"
	"travelando/internal/system/components"
	"travelando/internal/system/services"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if err := renderOwnerCalendar(w, r, http.StatusOK, "", "", false); err != nil {
		if errors.Is(err, services.ErrInvalidInput) {
			http.Error(w, "Mes o año inválido", http.StatusBadRequest)
			return
		}
		slog.Error("loading owner calendar", "error", err)
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func NewBlock(w http.ResponseWriter, r *http.Request) {
	view, err := ownerCalendarViewFromRequest(r)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	modal := ownerExternalBlockDialogEl(view.Cabins, nil, view.Month, view.Year, false)
	if hxhttp.IsRequest(r.Header) {
		if err := modal.Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Reserva externa | Travelando", Portal: system.PortalOwner, Yield: Group{ownerCalendarContentEl(view), modal}}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}

func renderOwnerCalendar(w http.ResponseWriter, r *http.Request, status int, flashLevel, flashMessage string, closeModal bool) error {
	view, err := ownerCalendarViewFromRequest(r)
	if err != nil {
		return err
	}
	content := ownerCalendarContentEl(view)
	if closeModal {
		content = Group{content, Div(ID("modal"), hx.SwapOOB("delete"))}
	}
	if flashMessage != "" {
		content = Group{content, Div(ID("global-flash-container"), hx.SwapOOB("true"), components.FlashToast(flashMessage, flashLevel))}
	}
	w.WriteHeader(status)
	if hxhttp.IsRequest(r.Header) {
		return content.Render(w)
	}
	return system.Layout{Title: "Mis cabañas | Travelando", Portal: system.PortalOwner, Yield: content}.New(r).Render(w)
}

func renderOwnerCalendarWithBlockModal(w http.ResponseWriter, r *http.Request, state *blockFormState, status int) error {
	view, err := ownerCalendarViewFromRequest(r)
	if err != nil {
		return err
	}
	content := ownerCalendarContentEl(view)
	modal := ownerExternalBlockDialogEl(view.Cabins, state, view.Month, view.Year, hxhttp.IsRequest(r.Header))
	w.WriteHeader(status)
	if hxhttp.IsRequest(r.Header) {
		return Group{content, modal}.Render(w)
	}
	return system.Layout{Title: "Mis cabañas | Travelando", Portal: system.PortalOwner, Yield: Group{content, modal}}.New(r).Render(w)
}

func ownerCalendarViewFromRequest(r *http.Request) (ownerCalendarView, error) {
	set := services.ServiceSetFromCtx(r.Context())
	if set.Calendar == nil {
		return ownerCalendarView{}, errors.New("calendar service unavailable")
	}
	now := time.Now()
	year, err := ownerCalendarParam(r, "year", now.Year())
	if err != nil {
		return ownerCalendarView{}, services.ErrInvalidInput
	}
	monthNumber, err := ownerCalendarParam(r, "month", int(now.Month()))
	if err != nil || monthNumber < 1 || monthNumber > 12 {
		return ownerCalendarView{}, services.ErrInvalidInput
	}
	data, err := set.Calendar.OwnerCalendar(ownerIDFromSession(r), year, time.Month(monthNumber))
	if err != nil {
		return ownerCalendarView{}, err
	}
	return newOwnerCalendarView(year, time.Month(monthNumber), data, now), nil
}

func ownerCalendarParam(r *http.Request, key string, fallback int) (int, error) {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback, nil
	}
	return strconv.Atoi(value)
}
