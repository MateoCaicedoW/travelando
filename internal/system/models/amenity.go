package models

type AmenityValueType string

const (
	AmenityBoolean  AmenityValueType = "boolean"
	AmenityQuantity AmenityValueType = "quantity"
)

type AmenityType struct {
	ID        int64
	OwnerID   int64
	Name      string
	ValueType AmenityValueType
}

type CabinAmenity struct {
	TypeID        int64
	Name          string
	ValueType     AmenityValueType
	BooleanValue  bool
	QuantityValue int
}
