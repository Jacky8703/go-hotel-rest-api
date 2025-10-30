package models

import "time"

type ServiceRequestDTO struct {
	ID         int    `json:"id,omitempty"`
	CustomerID int    `json:"customer_id" validate:"required"`
	ServiceID  int    `json:"service_id" validate:"required"`
	Date       string `json:"date" validate:"required,datetime=2006-01-02"`
}

type ServiceRequest struct {
	ID         int
	CustomerID int
	ServiceID  int
	Date       time.Time
}

type ServiceRequestPatch struct {
	CustomerID *int    `json:"customer_id,omitempty"`
	ServiceID  *int    `json:"service_id,omitempty"`
	Date       *string `json:"date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

const ServiceRequestValidationError = `Invalid service request data:
- Integer field 'customer_id' is required
- Integer field 'service_id' is required
- String field 'date' is required and must be in YYYY-MM-DD format`

func (s *ServiceRequest) ToDTO() ServiceRequestDTO {
	return ServiceRequestDTO{
		ID:         s.ID,
		CustomerID: s.CustomerID,
		ServiceID:  s.ServiceID,
		Date:       s.Date.Format("2006-01-02"),
	}
}

func (s *ServiceRequestDTO) ToModel() (ServiceRequest, error) {
	date, err := time.Parse("2006-01-02", s.Date)
	if err != nil {
		return ServiceRequest{}, err
	}
	return ServiceRequest{
		ID:         s.ID,
		CustomerID: s.CustomerID,
		ServiceID:  s.ServiceID,
		Date:       date,
	}, nil
}

func (p ServiceRequestPatch) FromStructToDBAttr() map[string]string {
	return map[string]string{
		"CustomerID": "customer_id",
		"ServiceID":  "service_id",
		"Date":       "service_date",
	}
}
