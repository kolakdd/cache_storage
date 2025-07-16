// Package database
package database

import (
	"database/sql"
	"log"

	"github.com/kolakdd/cache_storage/golang/repo"
)

func InitDB(envRepo repo.RepositoryEnv) (db *sql.DB) {
	dsn := envRepo.GetDatabaseDSN()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}
	configDB(db)
	err = db.Ping()
	if err != nil {
		log.Fatal("Error ping to database: ", err)
	}
	return db
}

func configDB(db *sql.DB) {
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
}
