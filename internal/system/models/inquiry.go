package models

const CurrentTermsDocumentSHA256 = "51394e505fdd87eab1cc5ed5ffca848bafdb690e735ad79b6fbac38c4cb8fde8"

type StayType string

const (
	StayDayUse    StayType = "day_use"
	StayOvernight StayType = "overnight"
)

type InquiryStatus string

const (
	InquiryPending   InquiryStatus = "pending"
	InquiryReserved  InquiryStatus = "reserved"
	InquiryLost      InquiryStatus = "lost"
	InquiryCancelled InquiryStatus = "cancelled"
)

type Inquiry struct {
	ID                    int64
	CabinID               int64
	CabinName             string
	OwnerName             string
	WhatsApp              string
	GuestName             string
	GuestCedula           string
	GuestPhone            string
	GuestEmail            string
	TermsAccepted         bool
	TermsAcceptedAt       string
	TermsDocumentSHA256   string
	StayType              StayType
	DateFrom              string
	DateTo                string
	StartTime             string
	EndTime               string
	CleaningHours         int
	Guests                int
	Estimated             int64
	DepositCOP            int64
	Status                InquiryStatus
	FinalTotal            int64
	CommissionBP          int
	Commission            int64
	CommissionPaidAt      string
	CommissionPaidAmount  int64
	CommissionBalance     int64
	CommissionAdjustments []CommissionAdjustment
	CreatedAt             string
	CancelledFromStatus   InquiryStatus
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

type InquiryUpdate struct {
	GuestName  string
	GuestPhone string
	GuestEmail string
	StayType   StayType
	DateFrom   string
	DateTo     string
	Guests     int
}
