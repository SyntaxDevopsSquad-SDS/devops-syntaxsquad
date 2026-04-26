package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://whoknows:whoknows@localhost:5432/whoknows?sslmode=disable"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Cannot connect to database: ", err)
	}

	schema := `
    DROP TABLE IF EXISTS users;
    CREATE TABLE IF NOT EXISTS users (
        id       SERIAL PRIMARY KEY,
        username TEXT NOT NULL UNIQUE,
        email    TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL
    );
    INSERT INTO users (username, email, password)
    VALUES ('admin', 'keamonk1@stud.kea.dk', '5f4dcc3b5aa765d61d8327deb882cf99')
    ON CONFLICT DO NOTHING;

    CREATE TABLE IF NOT EXISTS pages (
        title        TEXT PRIMARY KEY,
        url          TEXT NOT NULL UNIQUE,
        language     TEXT NOT NULL DEFAULT 'en' CHECK(language IN ('en', 'da')),
        last_updated TIMESTAMP,
        content      TEXT NOT NULL
    );`

	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Database initialized successfully")
}
