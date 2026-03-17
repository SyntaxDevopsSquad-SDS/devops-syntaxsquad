package main

import (
	"database/sql"
	"fmt"
	"log"
	_ "modernc.org/sqlite"
	"os"
)

// dbPath reads the database path from the DB_PATH environment variable,
// falling back to "whoknows.db" if not set.
func getDBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "whoknows.db"
}

// Global db variabel - kan bruges i alle filer
var db *sql.DB

func checkDBExists() bool {
	if _, err := os.Stat(getDBPath()); os.IsNotExist(err) {
		return false
	}
	return true
}

// connectDB initiates a connection, checks for file existence, and pings the database.
func connectDB() {
	if !checkDBExists() {
		fmt.Printf("Critical Error: Database file not found at %s\n", getDBPath())
		os.Exit(1)
	}

	var err error
	db, err = sql.Open("sqlite", getDBPath())
	if err != nil {
		fmt.Printf("could not open database: %v\n", err)
		os.Exit(1)
	}

	if err = db.Ping(); err != nil {
		fmt.Printf("database ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Connection Status: Successfully connected to", getDBPath())
}

// QueryDB executes a query and returns results as a slice of maps
func QueryDB(query string, args []interface{}, one bool) (interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("error closing rows: %v", err)
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	var results []map[string]interface{}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if one {
		if len(results) > 0 {
			return results[0], nil
		}
		return nil, nil
	}

	return results, nil
}
