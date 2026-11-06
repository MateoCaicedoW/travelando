package helpers

import (
	"net/http"

	hxhttp "maragu.dev/gomponents-htmx/http"
)

func Redirect(w http.ResponseWriter, r *http.Request, target string) {
	if hxhttp.IsRequest(r.Header) {
		hxhttp.SetRedirect(w.Header(), target)
		w.WriteHeader(http.StatusNoContent)
		return
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}
