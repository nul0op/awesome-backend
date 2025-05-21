package model

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB = nil

func GetLinks() (result []AwesomeLink) {

	select_statement := "select id, name from link"
	rows, err := db.Query(select_statement)

	if err != nil {
		fmt.Printf("SQL Error: %s\nreturning empty result set", err)
		return
	}

	for rows.Next() {
		awlink := AwesomeLink{}
		if err := rows.Scan(&awlink.OriginHash, &awlink.Name); err != nil {
			panic(err)
		}
		result = append(result, awlink)
	}

	return
}

func Connect() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	err := *new(error)
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("cannot connect to database with dsn [%s]\n", dsn)
	}

	// FIXME:
	// defer db.Close()
}
