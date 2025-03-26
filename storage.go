package main

import (
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func initDB() (*sql.DB, error) {
	dbPath := os.Getenv("TODO_DBFILE")
	if dbPath == "" {
		dbPath = "scheduler.db"
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date TEXT NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat TEXT
        );
        CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);
    `)
	if err != nil {
		return nil, err
	}

	return db, nil
}
