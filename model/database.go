package model

import (
	"awesome-portal/backend/util"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var db *sqlx.DB = nil

func GetLink(id string) (result []AwesomeLink) {

	result = []AwesomeLink{}

	sql := `
		select name, description, origin_url, level, subscribers_count, watchers_count, updated from link
	 	where external_id = $1`

	err := db.Select(&result, sql, id)
	if err != nil {
		util.Log.Errorf("SQL Error: %s. returning empty result set", err)
		return
	}
	return
}

func GetLinks(search string) (result []AwesomeLink) {

	result = []AwesomeLink{}

	if len(search) == 0 {
		search = "%"
	}

	sql := `
		select name, description, origin_url, level, subscribers_count, watchers_count, updated from link
	 	where name || ' ' || description ilike $1 limit 50`

	err := db.Select(&result, sql, "%"+search+"%")
	if err != nil {
		util.Log.Errorf("SQL Error: %s. returning empty result set", err)
		return
	}

	return
}

func trunc(s string, l int) string {
	if len(s) > l {
		return s[0:l]
	}
	return s
}

func SaveLink(link AwesomeLink) {
	_, err := db.Exec(
		`insert into link (
			external_id, level, name, description, origin_url, subscribers_count, watchers_count, topics, updated
		) values (
		 	$1, $2, $3, $4, $5, $6, $7, $8, $9
		)`,
		link.OriginHash, 0, link.Name, link.Description, link.OriginUrl,
		link.Subscribers, link.Watchers, trunc(link.Topics, 250), link.UpdateTs)

	if err != nil {
		util.Log.Errorf("SQL Error: %s.", err)
		return
	}
}

func Connect() {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASS")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	err := *new(error)
	db, err = sqlx.Open("postgres", dsn)
	if err != nil {
		util.Log.Errorf("unable to connect to database with dsn [%s]\n", dsn)
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		util.Log.Errorf("unable to connect to database on dsn [%s]: %v", dsn, err)
		panic("Exiting")
	}
	util.Log.Info("successfully connected to database")
	// FIXME:
	// defer db.Close()
}
