package home

import (
	"net/url"

	"travelando/internal/system/models"

	"go.leapkit.dev/core/form/validate"
)

type PageData struct {
	StayType      models.StayType
	DateFrom      string
	DateTo        string
	Guests        int
	Cabins        []models.Cabin
	DidSearch     bool
	Page          int
	HasNextPage   bool
	HasPrevPage   bool
	Error         string
	SearchValues  url.Values
	SearchErrors  validate.Errors
	RequestValues url.Values
	RequestErrors validate.Errors
}
