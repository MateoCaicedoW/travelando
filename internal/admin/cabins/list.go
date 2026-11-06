package cabins

import (
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"travelando/internal/system"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

const cabinsPerPage = 10

func Index(w http.ResponseWriter, r *http.Request) {
	svc := services.ServiceSetFromCtx(r.Context()).Cabins
	if svc == nil {
		http.Error(w, "Servicio no disponible", http.StatusInternalServerError)
		return
	}
	search := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(search) > 200 {
		http.Error(w, "Búsqueda demasiado larga", http.StatusBadRequest)
		return
	}
	pageNumber, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || pageNumber < 1 {
		pageNumber = 1
	}
	page, err := svc.SearchCabinsPage(search, pageNumber, cabinsPerPage)
	if err != nil {
		server.Error(w, err, http.StatusInternalServerError)
		return
	}
	if hxhttp.IsRequest(r.Header) {
		if err := cabinListResultsEl(page, search).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	pageEl := system.Layout{Title: "Cabañas | Travelando", Portal: system.PortalAdmin, Yield: listEl(page, search)}.New(r)
	if err := pageEl.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
