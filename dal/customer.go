package dal

import (
	"context"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllCustomers(ctx context.Context, conn *pgx.Conn) ([]models.Customer, error) {
	rows, _ := conn.Query(ctx, "SELECT id, cf, customer_name, age, email FROM customer")
	defer rows.Close()
	var customers []models.Customer
	for rows.Next() {
		var customer models.Customer
		err := rows.Scan(&customer.ID, &customer.CF, &customer.Name, &customer.Age, &customer.Email)
		if err != nil {
			return nil, err
		}
		customers = append(customers, customer)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return customers, nil
}

func GetCustomerByID(ctx context.Context, conn *pgx.Conn, customerID int) (*models.Customer, error) {
	row := conn.QueryRow(ctx, "SELECT id, cf, customer_name, age, email FROM customer WHERE id = $1", customerID)
	var customer models.Customer
	err := row.Scan(&customer.ID, &customer.CF, &customer.Name, &customer.Age, &customer.Email)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func CreateCustomer(ctx context.Context, conn *pgx.Conn, customer *models.Customer) error {
	row := conn.QueryRow(ctx, "INSERT INTO customer (cf, customer_name, age, email) VALUES ($1, $2, $3, $4) RETURNING id", customer.CF, customer.Name, customer.Age, customer.Email)
	err := row.Scan(&customer.ID)
	return err
}

func UpdateCustomerByID(ctx context.Context, conn *pgx.Conn, customer *models.Customer) error {
	row := conn.QueryRow(ctx, "UPDATE customer SET cf = $1, customer_name = $2, age = $3, email = $4 WHERE id = $5 RETURNING cf, customer_name, age, email", customer.CF, customer.Name, customer.Age, customer.Email, customer.ID)
	err := row.Scan(&customer.CF, &customer.Name, &customer.Age, &customer.Email)
	return err
}

func PatchCustomerByID(ctx context.Context, conn *pgx.Conn, customerID int, patch models.CustomerPatch) error {
	query, args := createPatchQuery("customer", patch, "id", customerID)
	tag, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func DeleteCustomerByID(ctx context.Context, conn *pgx.Conn, customerID int) error {
	tag, err := conn.Exec(ctx, "DELETE FROM customer WHERE id = $1", customerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
