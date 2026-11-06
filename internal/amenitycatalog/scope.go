package amenitycatalog

import "travelando/internal/system/models"

func managedTypes(types []models.AmenityType, ownerID int64, admin bool) []models.AmenityType {
	managed := make([]models.AmenityType, 0, len(types))
	for _, amenityType := range types {
		if admin && amenityType.OwnerID == 0 || !admin && amenityType.OwnerID == ownerID {
			managed = append(managed, amenityType)
		}
	}
	return managed
}

func scope(ownerID int64, admin bool) int64 {
	if admin {
		return 0
	}
	return ownerID
}
