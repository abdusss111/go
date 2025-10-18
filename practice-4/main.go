package main

import (
	"fmt"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// User struct represents a user in the database
type User struct {
	ID      int     `db:"id"`
	Name    string  `db:"name"`
	Email   string  `db:"email"`
	Balance float64 `db:"balance"`
}

// InsertUser inserts a new user into the database
func InsertUser(db *sqlx.DB, user User) error {
	query := `INSERT INTO users (name, email, balance) VALUES (:name, :email, :balance)`
	_, err := db.NamedExec(query, user)
	return err
}

// GetAllUsers retrieves all users from the database
func GetAllUsers(db *sqlx.DB) ([]User, error) {
	var users []User
	query := `SELECT id, name, email, balance FROM users ORDER BY id`
	err := db.Select(&users, query)
	return users, err
}

// GetUserByID retrieves a single user by their ID
func GetUserByID(db *sqlx.DB, id int) (User, error) {
	var user User
	query := `SELECT id, name, email, balance FROM users WHERE id = $1`
	err := db.Get(&user, query, id)
	return user, err
}

// TransferBalance transfers money from one user to another within a transaction
func TransferBalance(db *sqlx.DB, fromID int, toID int, amount float64) error {
	// Start a transaction
	tx, err := db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // This will be ignored if tx.Commit() is called

	// Check if sender has sufficient balance
	var senderBalance float64
	err = tx.Get(&senderBalance, "SELECT balance FROM users WHERE id = $1", fromID)
	if err != nil {
		return fmt.Errorf("failed to get sender balance: %w", err)
	}
	if senderBalance < amount {
		return fmt.Errorf("insufficient balance: sender has %.2f, trying to transfer %.2f", senderBalance, amount)
	}

	// Decrease the balance of the sender
	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", amount, fromID)
	if err != nil {
		return fmt.Errorf("failed to decrease sender balance: %w", err)
	}

	// Increase the balance for the receiver
	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, toID)
	if err != nil {
		return fmt.Errorf("failed to increase receiver balance: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func main() {
	// Database connection string
	dsn := "user=postgres password=password dbname=practice4_db sslmode=disable host=localhost port=5432"

	// Create database connection using sqlx
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Configure connection pooling
	db.SetMaxOpenConns(10)                // Maximum number of open connections
	db.SetMaxIdleConns(5)                 // Maximum number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Maximum lifetime for a connection

	// Test the connection
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("Successfully connected to PostgreSQL database!")

	// Example usage of CRUD operations

	// 1. Get all users
	fmt.Println("\n=== Getting all users ===")
	users, err := GetAllUsers(db)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
	} else {
		for _, user := range users {
			fmt.Printf("ID: %d, Name: %s, Email: %s, Balance: %.2f\n", 
				user.ID, user.Name, user.Email, user.Balance)
		}
	}

	// 2. Get a specific user by ID
	fmt.Println("\n=== Getting user by ID ===")
	user, err := GetUserByID(db, 1)
	if err != nil {
		log.Printf("Error getting user by ID: %v", err)
	} else {
		fmt.Printf("User found: ID: %d, Name: %s, Email: %s, Balance: %.2f\n", 
			user.ID, user.Name, user.Email, user.Balance)
	}

	// 3. Insert a new user
	fmt.Println("\n=== Inserting new user ===")
	newUser := User{
		Name:    "Charlie Wilson",
		Email:   "charlie.wilson@example.com",
		Balance: 500.00,
	}
	err = InsertUser(db, newUser)
	if err != nil {
		log.Printf("Error inserting user: %v", err)
	} else {
		fmt.Println("New user inserted successfully!")
	}

	// 4. Transfer balance between users
	fmt.Println("\n=== Transferring balance ===")
	fmt.Println("Before transfer:")
	user1, _ := GetUserByID(db, 1)
	user2, _ := GetUserByID(db, 2)
	fmt.Printf("User 1 balance: %.2f\n", user1.Balance)
	fmt.Printf("User 2 balance: %.2f\n", user2.Balance)

	transferAmount := 100.00
	err = TransferBalance(db, 1, 2, transferAmount)
	if err != nil {
		log.Printf("Error transferring balance: %v", err)
	} else {
		fmt.Printf("Successfully transferred %.2f from user 1 to user 2\n", transferAmount)
		
		fmt.Println("After transfer:")
		user1, _ = GetUserByID(db, 1)
		user2, _ = GetUserByID(db, 2)
		fmt.Printf("User 1 balance: %.2f\n", user1.Balance)
		fmt.Printf("User 2 balance: %.2f\n", user2.Balance)
	}

	// 5. Test transaction rollback (insufficient balance)
	fmt.Println("\n=== Testing transaction rollback (insufficient balance) ===")
	largeAmount := 10000.00
	err = TransferBalance(db, 1, 2, largeAmount)
	if err != nil {
		fmt.Printf("Transaction correctly failed: %v\n", err)
	} else {
		fmt.Println("Transaction should have failed but didn't!")
	}

	fmt.Println("\n=== Practice 4 completed successfully! ===")
}
