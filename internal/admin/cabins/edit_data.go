package cabins

import (
	"errors"

	"travelando/internal/system/models"
	"travelando/internal/system/services"
)

type cabinEditData struct {
	Cabin        models.Cabin
	Blocks       []models.CabinBlock
	Bookings     []models.Inquiry
	Owners       []models.Owner
	AmenityTypes []models.AmenityType
}

func loadCabinEditData(cabins services.CabinService, blocksSvc services.BlockService, inquiriesSvc services.InquiryService, ownersSvc services.OwnerService, amenitiesSvc services.AmenityService, id int64) (cabinEditData, error) {
	if cabins == nil || ownersSvc == nil || amenitiesSvc == nil {
		return cabinEditData{}, errors.New("cabin or owner service unavailable")
	}
	cabin, err := cabins.Cabin(id)
	if err != nil {
		return cabinEditData{}, err
	}
	if cabin == nil {
		return cabinEditData{}, services.ErrNotFound
	}
	data := cabinEditData{Cabin: *cabin}
	data.Owners, err = ownersSvc.Owners()
	if err != nil {
		return cabinEditData{}, err
	}
	data.AmenityTypes, err = amenitiesSvc.TypesForOwner(cabin.OwnerID)
	if err != nil {
		return cabinEditData{}, err
	}
	data.Blocks, data.Bookings, err = loadCabinFormOptions(blocksSvc, inquiriesSvc, id)
	return data, err
}

func loadCabinFormOptions(blocksSvc services.BlockService, inquiriesSvc services.InquiryService, id int64) ([]models.CabinBlock, []models.Inquiry, error) {
	if id == 0 {
		return nil, nil, nil
	}
	if blocksSvc == nil || inquiriesSvc == nil {
		return nil, nil, errors.New("calendar services unavailable")
	}
	blocks, err := blocksSvc.Blocks(id)
	if err != nil {
		return nil, nil, err
	}
	inquiries, err := inquiriesSvc.Inquiries()
	if err != nil {
		return nil, nil, err
	}
	bookings := make([]models.Inquiry, 0, len(inquiries))
	for _, inquiry := range inquiries {
		if inquiry.CabinID == id && inquiry.Status == models.InquiryReserved {
			bookings = append(bookings, inquiry)
		}
	}
	return blocks, bookings, nil
}
