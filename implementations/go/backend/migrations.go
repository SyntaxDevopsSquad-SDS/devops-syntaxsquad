package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runMigrations() error {
	// Try multiple paths (local dev and Docker)
	var migrationDir string
	possiblePaths := []string{
		"./migrations",
		"../migrations",
		"../../migrations",
		"/app/migrations",
		"implementations/go/migrations",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			migrationDir = path
			break
		}
	}

	if migrationDir == "" {
		log.Printf("warning: migrations directory not found, skipping migrations")
		return nil
	}

	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("could not read migrations directory %s: %w", migrationDir, err)
	}

	var files []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		filePath := filepath.Join(migrationDir, file.Name())
		sql, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("could not read migration file %s: %w", file.Name(), err)
		}

		statements := strings.Split(string(sql), ";")
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			if _, err := db.Exec(stmt); err != nil {
				log.Printf("warning: migration %s statement failed (might already exist): %v", file.Name(), err)
			}
		}

		fmt.Printf("✅ Migration executed: %s\n", file.Name())
	}

	return nil
}
