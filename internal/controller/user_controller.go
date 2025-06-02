package controller

import (
	"encoding/base64"
	"encoding/json"
	"kwan19961217/cursor-pagination/internal/domain/user"
	"net/http"
	"time"
)

type UserController struct {
	userService *user.UserService
}

func NewUserController(userService *user.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) ListUsers(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	nextCursor := queryParams.Get("next_cursor")
	start := queryParams.Get("start")
	end := queryParams.Get("end")
	order := queryParams.Get("order")

	// validate if query param pairs are correctly provided
	if nextCursor == "" && start == "" && end == "" && order == "" {
		http.Error(w, "either cursor or start, end and order is required", http.StatusBadRequest)
		return
	}

	if nextCursor != "" && (start != "" || end != "" || order != "") {
		http.Error(w, "when cursor is provided, start, end and order must not be provided", http.StatusBadRequest)
		return
	}

	if nextCursor == "" && (start == "" || end == "" || order == "") {
		http.Error(w, "when cursor is not provided, start, end and order must be provided", http.StatusBadRequest)
		return
	}

	// validate if received params are valid
	if nextCursor != "" {
		decodedCursor, err := base64.StdEncoding.DecodeString(nextCursor)
		if err != nil {
			http.Error(w, "failed to decode next_cursor", http.StatusBadRequest)
			return
		}

		var userCursor user.UserCursor
		err = json.Unmarshal(decodedCursor, &userCursor)
		if err != nil {
			http.Error(w, "failed to unmarshal next_cursor", http.StatusBadRequest)
			return
		}

		if userCursor.Order != "asc" && userCursor.Order != "desc" {
			http.Error(w, "invalid order format", http.StatusBadRequest)
			return
		}

		users, err := c.userService.ListUsers(userCursor.Start, userCursor.End, userCursor.Order, userCursor.UserId)
		if err != nil {
			http.Error(w, "failed to list users", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
		return
	}

	startTime, err := time.ParseInLocation(time.RFC3339, start, time.UTC)
	if err != nil {
		http.Error(w, "invalid start format", http.StatusBadRequest)
		return
	}

	endTime, err := time.ParseInLocation(time.RFC3339, end, time.UTC)
	if err != nil {
		http.Error(w, "invalid end format", http.StatusBadRequest)
		return
	}

	if order != "asc" && order != "desc" {
		http.Error(w, "invalid order format", http.StatusBadRequest)
		return
	}

	users, err := c.userService.ListUsers(startTime, endTime, order, "")
	if err != nil {
		http.Error(w, "failed to list users", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}
