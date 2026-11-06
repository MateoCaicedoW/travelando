package home

import (
	"log/slog"
	"net/http"

	"travelando/internal/system"
	"travelando/internal/system/helpers"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/server"
	hxhttp "maragu.dev/gomponents-htmx/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Solicitud inválida", http.StatusBadRequest)
		return
	}
	set := services.ServiceSetFromCtx(r.Context())
	data := searchPageData(r, set.Availability)
	if set.Cabins != nil {
		featuredCabin, err := set.Cabins.FeaturedCabin()
		if err != nil {
			slog.Error("loading featured cabin for home page", "error", err)
		} else {
			data.FeaturedCabin = featuredCabin
		}
		if !data.DidSearch {
			data.SampleCabins, err = set.Cabins.RandomPublicCabins()
			if err != nil {
				slog.Error("loading random cabins for home page", "error", err)
			}
		}
	}
	if len(data.Cabins) > 0 {
		if set.Deposits == nil {
			data.Error = "deposit service unavailable"
		} else if deposit, err := set.Deposits.DepositAmount(); err != nil {
			data.Error = err.Error()
		} else {
			data.DepositCOP = deposit
		}
	}
	if data.Error != "" {
		slog.Error("searching available cabins", "error", data.Error)
		if err := helpers.SetFlash(r, helpers.FlashError, "No se pudo cargar la disponibilidad. Intenta de nuevo."); err != nil {
			http.Error(w, "Sesión no disponible", http.StatusInternalServerError)
			return
		}
	}
	if hxhttp.IsRequest(r.Header) && !hxhttp.IsHistoryRestoreRequest(r.Header) {
		if err := IndexEl(data).Render(w); err != nil {
			server.Error(w, err, http.StatusInternalServerError)
		}
		return
	}
	page := system.Layout{Title: "Travelando | Tu próximo plan está más cerca", Description: "Encuentra cabañas disponibles para tu próxima escapada.", Yield: IndexEl(data)}.New(r)
	if err := page.Render(w); err != nil {
		server.Error(w, err, http.StatusInternalServerError)
	}
}
