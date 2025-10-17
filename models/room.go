package models

type Room struct {
	ID       int    `json:"id,omitempty"`
	Number   int    `json:"number" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=basic suite"`
	Price    int    `json:"price" validate:"required,gt=0"`
	Capacity int    `json:"capacity" validate:"required,gt=0"`
}

type RoomPatch struct {
	Number   *int    `json:"number,omitempty"`
	Type     *string `json:"type,omitempty" validate:"omitempty,oneof=basic suite"`
	Price    *int    `json:"price,omitempty" validate:"omitempty,gt=0"`
	Capacity *int    `json:"capacity,omitempty" validate:"omitempty,gt=0"`
}

func (p RoomPatch) FromStructToDBAttr() map[string]string {
	return map[string]string{
		"Number":   "number",
		"Type":     "type",
		"Price":    "price",
		"Capacity": "capacity",
	}
}

const RoomValidationError = `Invalid room data:
- Integer field 'number' is required
- String field 'type' is required and must be one of: 'base', 'suite'
- Integer field 'price' is required and must be greater than 0
- Integer field 'capacity' is required and must be greater than 0`
