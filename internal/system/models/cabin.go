package models

type Cabin struct {
	ID              int64
	OwnerID         int64
	OwnerName       string
	OwnerWhatsApp   string
	OwnerActive     bool
	Name            string
	Description     string
	Location        string
	LocationURL     string
	CoverImage      string
	IncludedGuests  int
	ExtraGuestPrice int64
	DayPrice        int64
	NightPrice      int64
	DayStart        string
	DayEnd          string
	CheckIn         string
	CheckOut        string
	Active          bool
}

type CabinBlock struct {
	ID       int64
	CabinID  int64
	DateFrom string
	DateTo   string
	Reason   string
	Source   string
}
