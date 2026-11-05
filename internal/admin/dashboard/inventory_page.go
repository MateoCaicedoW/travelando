package dashboard

import (
	"net/http"
	"strconv"
)

const cabinsPerPage = 5

func requestedCabinPage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("cabins_page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}
