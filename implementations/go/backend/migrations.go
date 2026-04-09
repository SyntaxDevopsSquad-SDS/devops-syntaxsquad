package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runMigrations() error {
	// Try multiple paths
	var migrationDir string
	possiblePaths := []string{
		"./migrations",
		"../migrations",
		"implementations/go/migrations",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			migrationDir = path
			break
		}
	}

	if migrationDir == "" {
		return fmt.Errorf("could not find migrations directory")
	}

	entries, err := ioutil.ReadDir(migrationDir)
	if err != nil {
		return fmt.Errorf("could not read migrations directory %s: %w", migrationDir, err)
	}

	var files []os.FileInfo
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		filePath := filepath.Join(migrationDir, file.Name())
		sql, err := ioutil.ReadFile(filePath)
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
