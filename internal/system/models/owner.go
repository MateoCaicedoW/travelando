package models

type Owner struct {
	ID               int64
	Name             string
	Cedula           string
	WhatsApp         string
	CabinCount       int
	CommissionPaid   int64
	CommissionDue    int64
	CommissionCredit int64
	HasPIN           bool
	Active           bool
	SessionVersion   int64
}
