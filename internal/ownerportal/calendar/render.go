package calendar

import (
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/components"

	. "maragu.dev/gomponents"
	hx "maragu.dev/gomponents-htmx"
	hxhttp "maragu.dev/gomponents-htmx/http"
	. "maragu.dev/gomponents/html"
)

func renderOwnerCalendar(w http.ResponseWriter, r *http.Request, status int, flashLevel, flashMessage string, closeModal bool) error {
	view, err := ownerCalendarViewFromRequest(r)
	if err != nil {
		return err
	}
	content := IndexContentEl(view)
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
	content := IndexContentEl(view)
	modal := BlockDialogEl(view.Cabins, state, view.Month, view.Year, hxhttp.IsRequest(r.Header))
	w.WriteHeader(status)
	if hxhttp.IsRequest(r.Header) {
		return Group{content, modal}.Render(w)
	}
	return system.Layout{Title: "Mis cabañas | Travelando", Portal: system.PortalOwner, Yield: Group{content, modal}}.New(r).Render(w)
}
