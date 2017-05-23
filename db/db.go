package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func GetDatabaseConnection(databaseName string, user string, password string) *sql.DB {
	psqlString := fmt.Sprintf("dbname=%s user=%s password=%s", databaseName, user, password)
	db, err := sql.Open("postgres", psqlString)

	if err != nil {
		log.Fatal(err)

	}

	return db
}
