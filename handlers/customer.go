package handlers

import (
	"encoding/json"
	"errors"
	"example/models"
	"example/services"
	"log"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

func GetAllCustomers(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customers, err := services.GetAllCustomers(r.Context(), dbConnection)
		if err != nil {
			http.Error(w, "Unable to get all customers", http.StatusServiceUnavailable)
			log.Println("Error getting customers:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, customers)
	}
}

func GetCustomerByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid customer ID", http.StatusBadRequest)
			return
		}
		customer, err := services.GetCustomerByID(r.Context(), dbConnection, customerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "customer not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to get customer", http.StatusServiceUnavailable)
			log.Println("Error getting customer:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, customer)
	}
}

func CreateCustomer(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newCustomer models.Customer
		err := json.NewDecoder(r.Body).Decode(&newCustomer)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(newCustomer)
		if err != nil {
			http.Error(w, models.CustomerValidationError, http.StatusBadRequest)
			return
		}
		err = services.CreateCustomer(r.Context(), dbConnection, &newCustomer)
		if err != nil {
			http.Error(w, "Unable to create customer", http.StatusServiceUnavailable)
			log.Println("Error creating customer:", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		returnJSON(w, newCustomer)
	}
}

func UpdateCustomerByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid customer ID", http.StatusBadRequest)
			return
		}
		var updatedCustomer models.Customer
		err = json.NewDecoder(r.Body).Decode(&updatedCustomer)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		updatedCustomer.ID = customerID
		err = validator.Struct(updatedCustomer)
		if err != nil {
			http.Error(w, models.CustomerValidationError, http.StatusBadRequest)
			return
		}
		status, err := services.UpdateCustomerByID(r.Context(), dbConnection, &updatedCustomer)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "customer not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to update customer", http.StatusServiceUnavailable)
			log.Println("Error updating customer:", err.Error())
			return
		}
		w.WriteHeader(status)
		returnJSON(w, updatedCustomer)
	}
}

func PatchCustomerByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid customer ID", http.StatusBadRequest)
			return
		}
		var patch models.CustomerPatch
		err = json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(patch)
		if err != nil {
			http.Error(w, "customer patch data is invalid", http.StatusBadRequest)
			return
		}
		err = services.PatchCustomerByID(r.Context(), dbConnection, customerID, patch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "customer not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to patch customer", http.StatusServiceUnavailable)
			log.Println("Error patching customer:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteCustomerByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid customer ID", http.StatusBadRequest)
			return
		}
		err = services.DeleteCustomerByID(r.Context(), dbConnection, customerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "customer not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to delete customer", http.StatusServiceUnavailable)
			log.Println("Error deleting customer:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func returnJSON(w http.ResponseWriter, payload any) {
	w.Header().Set("Content-Type", "application/json")
	bytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "Unable to marshal payload", http.StatusInternalServerError)
		log.Println("Error marshaling payload:", err.Error())
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		log.Println("Error writing response:", err.Error())
	}
}
