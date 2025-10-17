package models

type Customer struct {
	ID    int    `json:"id,omitempty"`
	CF    string `json:"cf" validate:"required"`
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age" validate:"required,gt=0"`
	Email string `json:"email" validate:"required,email"`
}

type CustomerPatch struct {
	CF    *string `json:"cf,omitempty"`
	Name  *string `json:"name,omitempty"`
	Age   *int    `json:"age,omitempty" validate:"omitempty,gt=0"`
	Email *string `json:"email,omitempty" validate:"omitempty,email"`
}

func (p CustomerPatch) FromStructToDBAttr() map[string]string {
	return map[string]string{
		"CF":    "cf",
		"Name":  "customer_name",
		"Age":   "age",
		"Email": "email",
	}
}

const CustomerValidationError = `Invalid customer data:
- String field 'cf' is required
- String field 'name' is required
- Integer field 'age' is required and must be greater than 0
- String field 'email' is required and must be a valid email address`
