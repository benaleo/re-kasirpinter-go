package main

import (
	"fmt"
	"os"

	"re-kasirpinter-go/config"

	"gorm.io/gorm"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		fmt.Printf("Warning: Failed to load environment variables: %v\n", err)
	}

	// Initialize database
	db, err := config.InitDb()
	if err != nil {
		fmt.Printf("Failed to initialize database: %v\n", err)
		os.Exit(1)
	}

	// Truncate all tables
	fmt.Println("Truncating all tables...")
	if err := truncateAllTables(db); err != nil {
		fmt.Printf("Failed to truncate tables: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All tables truncated successfully!")
	fmt.Println("Database is now fresh.")
}

func truncateAllTables(db *gorm.DB) error {
	// Disable foreign key checks to allow truncating tables with dependencies
	db.Exec("SET CONSTRAINTS ALL DEFERRED")
	db.Exec("SET session_replication_role = 'replica'")

	// Truncate tables in reverse order of dependencies
	// Many-to-many junction tables first, then dependent tables, then parent tables
	tables := []string{
		"blacklisted_tokens",
		"user_role_permissions",
		"users",
		"user_roles",
		"user_permissions",
	}

	for _, tableName := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName)).Error; err != nil {
			// If TRUNCATE fails, try DELETE
			if err := db.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error; err != nil {
				return fmt.Errorf("failed to delete from %s: %w", tableName, err)
			}
			// Reset sequence
			db.Exec(fmt.Sprintf("ALTER SEQUENCE %s_id_seq RESTART WITH 1", tableName))
		}
		fmt.Printf("Truncated table: %s\n", tableName)
	}

	// Re-enable foreign key checks
	db.Exec("SET session_replication_role = 'origin'")

	return nil
}
