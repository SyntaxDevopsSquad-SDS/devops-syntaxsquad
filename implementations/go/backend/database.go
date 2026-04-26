package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

func getDatabaseURL() string {
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}
	return "postgres://whoknows:whoknows@localhost:5432/whoknows?sslmode=disable"
}

func connectDB() {
	var err error
	db, err = sql.Open("postgres", getDatabaseURL())
	if err != nil {
		fmt.Printf("could not open database: %v\n", err)
		os.Exit(1)
	}
	if err = db.Ping(); err != nil {
		fmt.Printf("database ping failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Connection Status: Successfully connected to PostgreSQL")
}

func QueryDB(query string, args []interface{}, one bool) (interface{}, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Printf("error closing rows: %v\n", err)
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
