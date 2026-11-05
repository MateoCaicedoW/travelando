package services

import (
	"context"

	"go.leapkit.dev/core/db"
)

type serviceDB struct {
	conn db.ConnFn
}

const ServicesContextKey = "app_services"

type ServiceSet struct {
	Cabins       CabinService
	Availability AvailabilityService
	Inquiries    InquiryService
	Blocks       BlockService
	Calendar     CalendarService
	Commissions  CommissionService
	Deposits     DepositService
	Owners       OwnerService
	Amenities    AmenityService
}

type cabinService struct{ serviceDB }
type availabilityService struct{ serviceDB }
type inquiryService struct{ serviceDB }
type blockService struct{ serviceDB }
type calendarService struct{ serviceDB }
type commissionService struct{ serviceDB }
type depositService struct{ serviceDB }
type ownerService struct{ serviceDB }

func NewCabinService(conn db.ConnFn) CabinService { return &cabinService{serviceDB{conn}} }
func NewAvailabilityService(conn db.ConnFn) AvailabilityService {
	return &availabilityService{serviceDB{conn}}
}
func NewInquiryService(conn db.ConnFn) InquiryService   { return &inquiryService{serviceDB{conn}} }
func NewBlockService(conn db.ConnFn) BlockService       { return &blockService{serviceDB{conn}} }
func NewCalendarService(conn db.ConnFn) CalendarService { return &calendarService{serviceDB{conn}} }
func NewCommissionService(conn db.ConnFn) CommissionService {
	return &commissionService{serviceDB{conn}}
}
func NewDepositService(conn db.ConnFn) DepositService { return &depositService{serviceDB{conn}} }
func NewOwnerService(conn db.ConnFn) OwnerService     { return &ownerService{serviceDB{conn}} }
func NewAmenityService(conn db.ConnFn) AmenityService { return &amenityService{serviceDB{conn}} }

func NewServiceSet(conn db.ConnFn) ServiceSet {
	return ServiceSet{
		Cabins: NewCabinService(conn), Availability: NewAvailabilityService(conn),
		Inquiries: NewInquiryService(conn), Blocks: NewBlockService(conn),
		Calendar: NewCalendarService(conn), Commissions: NewCommissionService(conn), Deposits: NewDepositService(conn), Owners: NewOwnerService(conn),
		Amenities: NewAmenityService(conn),
	}
}

func ServiceSetFromCtx(ctx context.Context) ServiceSet {
	set, _ := ctx.Value(ServicesContextKey).(ServiceSet)
	return set
}
