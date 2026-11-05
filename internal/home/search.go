package home

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"travelando/internal/system/models"
	"travelando/internal/system/services"

	"go.leapkit.dev/core/form"
	"go.leapkit.dev/core/form/validate"
)

const cabinsPerPage = 9

func searchPageData(r *http.Request, svc services.AvailabilityService) PageData {
	data := PageData{StayType: models.StayDayUse, Guests: 2, SearchValues: r.Form, Page: 1}
	data.DateFrom = r.FormValue("date_from")
	if _, exists := r.Form["date_from"]; !exists {
		data.DateFrom = time.Now().Format(time.DateOnly)
	}
	data.DateTo = r.FormValue("date_to")
	if value := r.FormValue("stay_type"); value != "" {
		data.StayType = models.StayType(value)
	}
	if value := r.FormValue("guests"); value != "" {
		if guests, err := strconv.Atoi(value); err == nil {
			data.Guests = guests
		}
	}
	if page, err := strconv.Atoi(r.FormValue("page")); err == nil && page > 0 && page-1 <= int(^uint(0)>>1)/cabinsPerPage {
		data.Page = page
	}
	data.HasPrevPage = data.Page > 1
	data.SearchErrors = searchFormErrors(r)
	data.DidSearch = data.DateFrom != ""
	if data.SearchErrors.HasAny() || !data.DidSearch {
		return data
	}
	from, err := time.Parse(time.DateOnly, data.DateFrom)
	if err != nil {
		return data
	}
	offset := (data.Page - 1) * cabinsPerPage
	switch data.StayType {
	case models.StayOvernight:
		to, err := time.Parse(time.DateOnly, data.DateTo)
		if err != nil {
			return data
		}
		result, searchError := searchCabins(svc, services.SearchInput{StayType: data.StayType, DateFrom: from, DateTo: to, Limit: cabinsPerPage, Offset: offset})
		data.Cabins, data.Error = result.Cabins, searchError
		data.HasNextPage = result.Total > offset && result.Total-offset > len(result.Cabins)
	case models.StayDayUse:
		data.DateTo = data.DateFrom
		result, searchError := searchCabins(svc, services.SearchInput{StayType: data.StayType, DateFrom: from, DateTo: from, Limit: cabinsPerPage, Offset: offset})
		data.Cabins, data.Error = result.Cabins, searchError
		data.HasNextPage = result.Total > offset && result.Total-offset > len(result.Cabins)
	}
	return data
}

func searchCabins(svc services.AvailabilityService, input services.SearchInput) (services.CabinSearchResult, string) {
	if svc == nil {
		return services.CabinSearchResult{}, "availability service unavailable"
	}
	result, err := svc.Search(input)
	if err != nil {
		return services.CabinSearchResult{}, err.Error()
	}
	return result, ""
}

func searchFormErrors(r *http.Request) validate.Errors {
	stayType := models.StayType(r.FormValue("stay_type"))
	dateFrom := r.FormValue("date_from")
	return form.Validate(r, validate.Fields(
		validate.Field("stay_type", validSearchStayType),
		validate.Field("date_from", validSearchStartDate),
		validate.Field("date_to", validSearchEndDate(stayType, dateFrom)),
		validate.Field("guests", validSearchGuestCount),
	))
}

func validSearchStayType(values []string) error {
	if len(values) == 0 || values[0] == "" || values[0] == string(models.StayDayUse) || values[0] == string(models.StayOvernight) {
		return nil
	}
	return errors.New("Selecciona pasadía o estadía.")
}

func validSearchStartDate(values []string) error {
	if len(values) == 0 || values[0] == "" {
		return nil
	}
	if _, err := time.Parse(time.DateOnly, values[0]); err != nil || values[0] < time.Now().Format(time.DateOnly) {
		return errors.New("Selecciona una fecha válida desde hoy en adelante.")
	}
	return nil
}

func validSearchEndDate(stayType models.StayType, dateFrom string) validate.ValidatorFn {
	return func(values []string) error {
		if dateFrom == "" || stayType != models.StayOvernight {
			return nil
		}
		if len(values) == 0 || values[0] == "" {
			return errors.New("Selecciona la fecha de salida.")
		}
		from, fromErr := time.Parse(time.DateOnly, dateFrom)
		to, toErr := time.Parse(time.DateOnly, values[0])
		if fromErr != nil || toErr != nil || !to.After(from) || to.Sub(from).Hours()/24 > 365 {
			return errors.New("La salida debe ser posterior a la llegada y estar dentro de un año.")
		}
		return nil
	}
}

func validSearchGuestCount(values []string) error {
	if len(values) == 0 || values[0] == "" {
		return nil
	}
	guests, err := strconv.Atoi(values[0])
	if err != nil || guests < 1 {
		return errors.New("Indica una cantidad válida de huéspedes.")
	}
	return nil
}
