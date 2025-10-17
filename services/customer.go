package services

import (
	"context"
	"example/dal"
	"example/models"

	"github.com/jackc/pgx/v5"
)

func GetAllCustomers(ctx context.Context, conn *pgx.Conn) ([]models.Customer, error) {
	return dal.GetAllCustomers(ctx, conn)
}

func GetCustomerByID(ctx context.Context, conn *pgx.Conn, customerID int) (*models.Customer, error) {
	return dal.GetCustomerByID(ctx, conn, customerID)
}

func CreateCustomer(ctx context.Context, conn *pgx.Conn, customer *models.Customer) error {
	return dal.CreateCustomer(ctx, conn, customer)
}

func UpdateCustomerByID(ctx context.Context, conn *pgx.Conn, customer *models.Customer) error {
	return dal.UpdateCustomerByID(ctx, conn, customer)
}

func PatchCustomerByID(ctx context.Context, conn *pgx.Conn, customerID int, patch models.CustomerPatch) error {
	return dal.PatchCustomerByID(ctx, conn, customerID, patch)
}

func DeleteCustomerByID(ctx context.Context, conn *pgx.Conn, customerID int) error {
	return dal.DeleteCustomerByID(ctx, conn, customerID)
}
