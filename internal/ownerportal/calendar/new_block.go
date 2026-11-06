package calendar

import (
	"net/http"

	"travelando/internal/system"

	"go.leapkit.dev/core/server"
	. "maragu.dev/gomponents"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

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
