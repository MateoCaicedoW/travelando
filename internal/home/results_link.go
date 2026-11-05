package home

import (
	"fmt"
	"net/url"
	"strconv"
)

func resultsLink(cabinID int64, data PageData) string {
	return fmt.Sprintf("/cabins/%d?stay_type=%s&date_from=%s&date_to=%s&guests=%d", cabinID, data.StayType, data.DateFrom, data.DateTo, data.Guests)
}

func resultsPageLink(data PageData, page int) string {
	query := url.Values{
		"stay_type": {string(data.StayType)},
		"date_from": {data.DateFrom},
		"date_to":   {data.DateTo},
		"guests":    {strconv.Itoa(data.Guests)},
		"page":      {strconv.Itoa(page)},
	}
	return "/?" + query.Encode()
}
