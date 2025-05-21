# awesome-portal-indexer
Awesome portal indexer/crawler in GO


## Setup
- create a database
postgres=# create database awesome_portal_db_go

- To upgrade
$HOME/go/bin/migrate -path db/migrations -database "postgres://postgres:example@127.0.0.1/awesome_portal_db_go?sslmode=disable" up


## Tools used:
- https://github.com/golang-migrate/migrate
- https://github.com/dal-go ??
