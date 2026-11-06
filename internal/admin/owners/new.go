package owners

import (
	"net/http"

	"travelando/internal/system"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func New(w http.ResponseWriter, r *http.Request) {
	dialog := ownerCreateDialogEl(nil, false)
	if hxhttp.IsRequest(r.Header) {
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
