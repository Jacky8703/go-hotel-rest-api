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

func GetAllRooms(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rooms, err := services.GetAllRooms(r.Context(), dbConnection)
		if err != nil {
			http.Error(w, "Unable to get all rooms", http.StatusServiceUnavailable)
			log.Println("Error getting rooms:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, rooms)
	}
}

func GetRoomByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}
		room, err := services.GetRoomByID(r.Context(), dbConnection, roomID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Room not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to get room", http.StatusServiceUnavailable)
			log.Println("Error getting room:", err.Error())
			return
		}
		w.WriteHeader(http.StatusOK)
		returnJSON(w, room)
	}
}

func CreateRoom(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newRoom models.Room
		err := json.NewDecoder(r.Body).Decode(&newRoom)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(newRoom)
		if err != nil {
			http.Error(w, models.RoomValidationError, http.StatusBadRequest)
			return
		}
		err = services.CreateRoom(r.Context(), dbConnection, &newRoom)
		if err != nil {
			http.Error(w, "Unable to create room", http.StatusServiceUnavailable)
			log.Println("Error creating room:", err.Error())
			return
		}
		w.WriteHeader(http.StatusCreated)
		returnJSON(w, newRoom)
	}
}

func UpdateRoomByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}
		var updatedRoom models.Room
		err = json.NewDecoder(r.Body).Decode(&updatedRoom)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		updatedRoom.ID = roomID
		err = validator.Struct(updatedRoom)
		if err != nil {
			http.Error(w, models.RoomValidationError, http.StatusBadRequest)
			return
		}
		status, err := services.UpdateRoomByID(r.Context(), dbConnection, &updatedRoom)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Room not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to update the room", http.StatusServiceUnavailable)
			log.Println("Error updating room:", err.Error())
			return
		}
		w.WriteHeader(status)
		returnJSON(w, updatedRoom)
	}
}

func PatchRoomByID(dbConnection *pgx.Conn, validator *validator.Validate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}
		var patch models.RoomPatch
		err = json.NewDecoder(r.Body).Decode(&patch)
		if err != nil {
			http.Error(w, "Invalid JSON format", http.StatusBadRequest)
			return
		}
		err = validator.Struct(patch)
		if err != nil {
			http.Error(w, "Room patch data is invalid", http.StatusBadRequest)
			return
		}
		err = services.PatchRoomByID(r.Context(), dbConnection, roomID, patch)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Room not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to patch room", http.StatusServiceUnavailable)
			log.Println("Error patching room:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func DeleteRoomByID(dbConnection *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roomID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid room ID", http.StatusBadRequest)
			return
		}
		err = services.DeleteRoomByID(r.Context(), dbConnection, roomID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				http.Error(w, "Room not found", http.StatusNotFound)
				return
			}
			http.Error(w, "Unable to delete room", http.StatusServiceUnavailable)
			log.Println("Error deleting room:", err.Error())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
