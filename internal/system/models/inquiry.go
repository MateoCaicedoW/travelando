package models

type StayType string

const (
	StayDayUse    StayType = "day_use"
	StayOvernight StayType = "overnight"
)

type InquiryStatus string

const (
	InquiryPending  InquiryStatus = "pending"
	InquiryReserved InquiryStatus = "reserved"
	InquiryLost     InquiryStatus = "lost"
)

type Inquiry struct {
	ID                    int64
	CabinID               int64
	CabinName             string
	OwnerName             string
	WhatsApp              string
	GuestName             string
	GuestPhone            string
	GuestEmail            string
	StayType              StayType
	DateFrom              string
	DateTo                string
	StartTime             string
	EndTime               string
	CleaningHours         int
	Guests                int
	Estimated             int64
	Status                InquiryStatus
	FinalTotal            int64
	CommissionBP          int
	Commission            int64
	CommissionPaidAt      string
	CommissionPaidAmount  int64
	CommissionBalance     int64
	CommissionAdjustments []CommissionAdjustment
	CreatedAt             string
}

type CommissionAdjustment struct {
	PreviousAmount int64
	NewAmount      int64
	Amount         int64
	ActorRole      string
	CreatedAt      string
}

type ReservationUpdate struct {
	GuestName     string
	GuestPhone    string
	GuestEmail    string
	StayType      StayType
	DateFrom      string
	DateTo        string
	Guests        int
	FinalTotal    int64
	UseFinalTotal bool
}
