package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"tiny-http/internal/middleware"
)

// JSONError helper for consistent API errors
func JSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// GET /user?id=123
func getUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		JSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"user_id": id})
}

// POST /user {"name":"Alice"}
func postUserHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid name")
		return
	}
	if body.Name == "" {
		JSONError(w, http.StatusBadRequest, "invalid name")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"created": body.Name})
}

func main() {
	mux := http.NewServeMux()

	// Protect both routes with middleware
	mux.Handle("/user", middleware.APIKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getUserHandler(w, r)
		case http.MethodPost:
			postUserHandler(w, r)
		default:
			JSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})))

	fmt.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

