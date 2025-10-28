package models

type HotelService struct {
	ID          int    `json:"id,omitempty"`
	Type        string `json:"type" validate:"required,oneof=cleaning room_service massage"`
	Description string `json:"description" validate:"required"`
	Duration    int    `json:"duration" validate:"required,min=1"` // number of minutes
}

type HotelServicePatch struct {
	Type        *string `json:"type,omitempty" validate:"omitempty,oneof=cleaning room_service massage"`
	Description *string `json:"description,omitempty"`
	Duration    *int    `json:"duration,omitempty" validate:"omitempty,min=1"`
}

const HotelServiceValidationError = `Invalid hotel service data:
- String field 'type' is required and must be one of: cleaning, room_service, massage
- String field 'description' is required
- Integer field 'duration' is required and must be greater than 0, representing minutes`

func (p HotelServicePatch) FromStructToDBAttr() map[string]string {
	return map[string]string{
		"Type":        "service_type",
		"Description": "description",
		"Duration":    "duration",
	}
}
