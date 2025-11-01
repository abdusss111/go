package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type User struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	TotalOrders int    `json:"total_orders"`
}

var db *sql.DB

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, r.URL.RawQuery)
		
		next(w, r)
		
		duration := time.Since(start)
		log.Printf("[%s] %s - Completed in %v", r.Method, r.URL.Path, duration)
	}
}

func main() {
	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL environment variable not set")
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/users", loggingMiddleware(getUsersHandler))

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	city := r.URL.Query().Get("city")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	query := `
		SELECT u.id, u.name, u.city, COUNT(o.id) AS total_orders
		FROM users u
		LEFT JOIN orders o ON u.id = o.user_id
	`

	var conditions []string
	var args []interface{}
	argCount := 1

	if city != "" {
		conditions = append(conditions, fmt.Sprintf("u.city = $%d", argCount))
		args = append(args, city)
		argCount++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " GROUP BY u.id, u.name, u.city"
	query += " ORDER BY total_orders DESC, u.id DESC"

	if limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			query += fmt.Sprintf(" LIMIT $%d", argCount)
			args = append(args, limit)
			argCount++
		}
	}

	if offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, offset)
		}
	}

	start := time.Now()
	rows, err := db.Query(query, args...)
	queryDuration := time.Since(start)

	if err != nil {
		log.Printf("Query error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.City, &u.TotalOrders); err != nil {
			log.Printf("Scan error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		users = append(users, u)
	}

	if err = rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []User{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Query-Time", fmt.Sprintf("%.3fms", float64(queryDuration.Microseconds())/1000.0))

	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("JSON encode error: %v", err)
	}
}