package services

import (
	"context"
	"example/dal"
	"example/models"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func GetAllHotelServices(ctx context.Context, conn *pgx.Conn) ([]models.HotelService, error) {
	return dal.GetAllHotelServices(ctx, conn)
}

func GetHotelServiceByID(ctx context.Context, conn *pgx.Conn, serviceID int) (*models.HotelService, error) {
	return dal.GetHotelServiceByID(ctx, conn, serviceID)
}

func CreateHotelService(ctx context.Context, conn *pgx.Conn, service *models.HotelService) error {
	err := validateHotelService(ctx, conn, service)
	if err != nil {
		return err
	}
	return dal.CreateHotelService(ctx, conn, service)
}

func UpdateHotelServiceByID(ctx context.Context, conn *pgx.Conn, service *models.HotelService) (int, error) {
	err := validateHotelService(ctx, conn, service)
	if err != nil {
		return 0, err
	}
	_, err = dal.GetHotelServiceByID(ctx, conn, service.ID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return http.StatusCreated, dal.CreateHotelService(ctx, conn, service)
		}
		return 0, err
	}
	return http.StatusOK, dal.UpdateHotelServiceByID(ctx, conn, service)
}

func PatchHotelServiceByID(ctx context.Context, conn *pgx.Conn, serviceID int, patch models.HotelServicePatch) error {
	// first check that the patch is valid
	oldService, err := dal.GetHotelServiceByID(ctx, conn, serviceID)
	if err != nil {
		return err
	}

	if patch.Type != nil {
		oldService.Type = *patch.Type
	}
	if patch.Description != nil {
		oldService.Description = *patch.Description
	}
	if patch.Duration != nil {
		oldService.Duration = *patch.Duration
	}

	err = validateHotelService(ctx, conn, oldService)
	if err != nil {
		return err
	}

	return dal.PatchHotelServiceByID(ctx, conn, serviceID, patch)
}

func DeleteHotelServiceByID(ctx context.Context, conn *pgx.Conn, serviceID int) error {
	return dal.DeleteHotelServiceByID(ctx, conn, serviceID)
}

func validateHotelService(ctx context.Context, conn *pgx.Conn, service *models.HotelService) error {
	services, err := dal.GetAllHotelServices(ctx, conn)
	if err != nil {
		return err
	}
	// check if the service already exists
	for _, s := range services {
		if s.Type == service.Type && s.ID != service.ID {
			return models.ValidationError{Message: fmt.Sprintf("Service of type %s already exists", s.Type)}
		}
	}
	return nil
}
