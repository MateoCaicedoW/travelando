package cabins

import (
	"net/http"
	"strconv"

	"travelando/internal/system/services"
)

func ownerCabinPageParam(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func ownerCabinsURL(page int) string { return "/owner/cabins?page=" + strconv.Itoa(page) }

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
