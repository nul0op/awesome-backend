package model

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "example"
	dbname   = "awesome_portal_db_go"
)

func Connect() {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		fmt.Printf("cannot connect to database with dsn [%s]\n", dsn)
	}

	select_statement := "select id, name from link"
	rows, err := db.Query(select_statement)
	if err != nil {
		fmt.Printf("SQL Error !: %s", err)
	}
	defer db.Close()

	for rows.Next() {
		var (
			id   int
			name string
		)
		if err := rows.Scan(&id, &name); err != nil {
			panic(err)
		}
		fmt.Printf("%s is %d\n", name, id)
	}

	// for v, i := range rows {
	// 	fmt.Println(v)
	// }
}
