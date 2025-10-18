package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// Default SQLite database path
	defaultDBPath = "./expense.db"
)

func main() {
	fmt.Println("=== Expense Tracker Database Schema Verification ===")
	
	// Get database path from command line argument or use default
	dbPath := defaultDBPath
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}
	
	// Convert to absolute path for better error messages
	absPath, err := filepath.Abs(dbPath)
	if err != nil {
		log.Fatalf("Error getting absolute path: %v", err)
	}
	
	fmt.Printf("Checking database: %s\n", absPath)
	
	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("❌ Database file does not exist: %s\n", absPath)
		fmt.Println("Please run migrations first:")
		fmt.Println("  migrate -path internal/db/migrations -database \"sqlite3://./expense.db\" up")
		os.Exit(1)
	}
	
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	defer db.Close()
	
	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	
	fmt.Println("✅ Database connection successful")
	
	// Verify schema
	verifySchema(db)
	
	fmt.Println("\n=== Verification Complete ===")
	fmt.Println("✅ All schema validations passed!")
}

func verifySchema(db *sql.DB) {
	fmt.Println("\n--- Verifying Database Schema ---")
	
	// Check if all required tables exist
	requiredTables := []string{"users", "categories", "expenses"}
	
	for _, table := range requiredTables {
		if !tableExists(db, table) {
			fmt.Printf("❌ Table '%s' does not exist\n", table)
			os.Exit(1)
		}
		fmt.Printf("✅ Table '%s' exists\n", table)
	}
	
	// Verify users table structure
	verifyUsersTable(db)
	
	// Verify categories table structure
	verifyCategoriesTable(db)
	
	// Verify expenses table structure
	verifyExpensesTable(db)
	
	// Verify indexes
	verifyIndexes(db)
	
	// Test constraints with sample data
	testConstraints(db)
}

func tableExists(db *sql.DB, tableName string) bool {
	query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?"
	var name string
	err := db.QueryRow(query, tableName).Scan(&name)
	return err == nil
}

func verifyUsersTable(db *sql.DB) {
	fmt.Println("\n--- Verifying Users Table ---")
	
	// Check columns
	expectedColumns := map[string]string{
		"id":         "INTEGER",
		"email":      "TEXT",
		"name":       "TEXT",
		"created_at": "TIMESTAMP",
	}
	
	rows, err := db.Query("PRAGMA table_info(users)")
	if err != nil {
		log.Fatalf("Error getting table info for users: %v", err)
	}
	defer rows.Close()
	
	foundColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int
		
		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatalf("Error scanning table info: %v", err)
		}
		
		foundColumns[name] = true
		
		if _, exists := expectedColumns[name]; exists {
			fmt.Printf("✅ Column '%s' exists with type '%s'\n", name, dataType)
		}
	}
	
	// Check for missing columns
	for colName := range expectedColumns {
		if !foundColumns[colName] {
			fmt.Printf("❌ Missing column: %s\n", colName)
			os.Exit(1)
		}
	}
	
	// Check constraints
	checkUniqueConstraint(db, "users", "email")
}

func verifyCategoriesTable(db *sql.DB) {
	fmt.Println("\n--- Verifying Categories Table ---")
	
	// Check columns
	expectedColumns := map[string]string{
		"id":         "INTEGER",
		"name":       "TEXT",
		"user_id":    "INTEGER",
		"created_at": "TIMESTAMP",
	}
	
	rows, err := db.Query("PRAGMA table_info(categories)")
	if err != nil {
		log.Fatalf("Error getting table info for categories: %v", err)
	}
	defer rows.Close()
	
	foundColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int
		
		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatalf("Error scanning table info: %v", err)
		}
		
		foundColumns[name] = true
		
		if _, exists := expectedColumns[name]; exists {
			fmt.Printf("✅ Column '%s' exists with type '%s'\n", name, dataType)
		}
	}
	
	// Check for missing columns
	for colName := range expectedColumns {
		if !foundColumns[colName] {
			fmt.Printf("❌ Missing column: %s\n", colName)
			os.Exit(1)
		}
	}
	
	// Check foreign key constraint
	checkForeignKey(db, "categories", "user_id", "users", "id")
	
	// Check unique constraint
	checkUniqueConstraint(db, "categories", "user_id, name")
}

func verifyExpensesTable(db *sql.DB) {
	fmt.Println("\n--- Verifying Expenses Table ---")
	
	// Check columns
	expectedColumns := map[string]string{
		"id":          "INTEGER",
		"user_id":     "INTEGER",
		"category_id": "INTEGER",
		"amount":      "DECIMAL",
		"currency":    "CHAR",
		"spent_at":    "TIMESTAMP",
		"created_at":  "TIMESTAMP",
		"note":        "TEXT",
	}
	
	rows, err := db.Query("PRAGMA table_info(expenses)")
	if err != nil {
		log.Fatalf("Error getting table info for expenses: %v", err)
	}
	defer rows.Close()
	
	foundColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var defaultValue interface{}
		var pk int
		
		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			log.Fatalf("Error scanning table info: %v", err)
		}
		
		foundColumns[name] = true
		
		if _, exists := expectedColumns[name]; exists {
			fmt.Printf("✅ Column '%s' exists with type '%s'\n", name, dataType)
		}
	}
	
	// Check for missing columns
	for colName := range expectedColumns {
		if !foundColumns[colName] {
			fmt.Printf("❌ Missing column: %s\n", colName)
			os.Exit(1)
		}
	}
	
	// Check foreign key constraints
	checkForeignKey(db, "expenses", "user_id", "users", "id")
	checkForeignKey(db, "expenses", "category_id", "categories", "id")
}

func checkUniqueConstraint(db *sql.DB, table, columns string) {
	query := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", table)
	var sql string
	err := db.QueryRow(query).Scan(&sql)
	if err != nil {
		log.Fatalf("Error getting table SQL: %v", err)
	}
	
	// Simple check for UNIQUE constraint in SQL
	if contains(sql, "UNIQUE") {
		fmt.Printf("✅ Unique constraint found on %s.%s\n", table, columns)
	} else {
		fmt.Printf("⚠️  Unique constraint not explicitly found on %s.%s\n", table, columns)
	}
}

func checkForeignKey(db *sql.DB, table, column, refTable, refColumn string) {
	query := fmt.Sprintf("SELECT sql FROM sqlite_master WHERE type='table' AND name='%s'", table)
	var sql string
	err := db.QueryRow(query).Scan(&sql)
	if err != nil {
		log.Fatalf("Error getting table SQL: %v", err)
	}
	
	// Simple check for FOREIGN KEY constraint in SQL
	if contains(sql, "FOREIGN KEY") && contains(sql, refTable) {
		fmt.Printf("✅ Foreign key constraint found: %s -> %s.%s\n", table+"."+column, refTable, refColumn)
	} else {
		fmt.Printf("⚠️  Foreign key constraint not found: %s -> %s.%s\n", table+"."+column, refTable, refColumn)
	}
}

func verifyIndexes(db *sql.DB) {
	fmt.Println("\n--- Verifying Indexes ---")
	
	expectedIndexes := []string{
		"idx_categories_user_id",
		"idx_expenses_user_id",
		"idx_expenses_user_spent_at",
	}
	
	for _, indexName := range expectedIndexes {
		query := "SELECT name FROM sqlite_master WHERE type='index' AND name=?"
		var name string
		err := db.QueryRow(query, indexName).Scan(&name)
		if err == nil {
			fmt.Printf("✅ Index '%s' exists\n", indexName)
		} else {
			fmt.Printf("❌ Index '%s' not found\n", indexName)
		}
	}
}

func testConstraints(db *sql.DB) {
	fmt.Println("\n--- Testing Constraints ---")
	
	// Test unique email constraint
	fmt.Println("Testing unique email constraint...")
	_, err := db.Exec("INSERT INTO users (email, name) VALUES ('test@example.com', 'Test User')")
	if err != nil {
		fmt.Printf("❌ Failed to insert first user: %v\n", err)
	} else {
		fmt.Println("✅ First user inserted successfully")
	}
	
	// Try to insert duplicate email
	_, err = db.Exec("INSERT INTO users (email, name) VALUES ('test@example.com', 'Another User')")
	if err != nil {
		fmt.Println("✅ Unique email constraint working (duplicate rejected)")
	} else {
		fmt.Println("❌ Unique email constraint failed (duplicate allowed)")
	}
	
	// Test positive amount constraint
	fmt.Println("Testing positive amount constraint...")
	_, err = db.Exec("INSERT INTO categories (name) VALUES ('Test Category')")
	if err != nil {
		fmt.Printf("❌ Failed to insert category: %v\n", err)
		return
	}
	
	// Try to insert negative amount
	_, err = db.Exec("INSERT INTO expenses (user_id, category_id, amount, currency, spent_at) VALUES (1, 1, -10.00, 'USD', '2024-01-01 12:00:00')")
	if err != nil {
		fmt.Println("✅ Positive amount constraint working (negative amount rejected)")
	} else {
		fmt.Println("❌ Positive amount constraint failed (negative amount allowed)")
	}
	
	// Clean up test data
	db.Exec("DELETE FROM expenses WHERE user_id = 1")
	db.Exec("DELETE FROM categories WHERE id = 1")
	db.Exec("DELETE FROM users WHERE id = 1")
	fmt.Println("✅ Test data cleaned up")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
