package model

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB = nil

func GetLinks(search string) (result []AwesomeLink) {

	select_statement := "select name, description, origin_url, level, subscribers_count, watchers_count, updated from link"
	if len(search) > 0 {
		select_statement = select_statement + fmt.Sprintf(" where name like '%%%s%%'", search)
	}

	rows, err := db.Query(select_statement)

	if err != nil {
		Log.Errorf("SQL Error: %s. returning empty result set", err)
		return
	}

	for rows.Next() {
		awlink := AwesomeLink{}
		if err := rows.Scan(
			&awlink.Name,
			&awlink.Description,
			&awlink.OriginUrl,
			&awlink.Level,
			&awlink.Subscribers,
			&awlink.Watchers,
			&awlink.UpdateTs,
		); err != nil {
			panic(err)
		}
		result = append(result, awlink)
	}

	return
}

func trunc(s string, l int) string {
	if len(s) > l {
		return s[0:l]
	}
	return s
}

func SaveLinks(link AwesomeLink) {

	_, err := db.Exec(
		`insert into link (
			external_id, level, name, description, origin_url, subscribers_count, watchers_count, topics
		) values (
		 	$1, $2, $3, $4, $5, $6, $7, $8
		)`,
		link.OriginHash, 0, link.Name, link.Description, link.OriginUrl, link.Subscribers, link.Watchers, trunc(link.Topics, 250))

	if err != nil {
		Log.Errorf("SQL Error: %s. returning empty result set", err)
		return
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
		Log.Errorf("unable to connect to database with dsn [%s]\n", dsn)
	}

	// FIXME:
	// defer db.Close()
}
