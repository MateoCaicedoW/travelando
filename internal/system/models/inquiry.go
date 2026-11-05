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
	ID               int64
	CabinID          int64
	CabinName        string
	OwnerName        string
	WhatsApp         string
	GuestName        string
	GuestPhone       string
	GuestEmail       string
	StayType         StayType
	DateFrom         string
	DateTo           string
	Guests           int
	Estimated        int64
	Status           InquiryStatus
	FinalTotal       int64
	CommissionBP     int
	Commission       int64
	CommissionPaidAt string
	CreatedAt        string
}
