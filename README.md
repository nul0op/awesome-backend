# awesome-portal-indexer
Awesome portal indexer/crawler in GO

Setup
- create a database

➜  awesome-backend git:(main U:2 ?:3) ✗ docker run -v ./scripts:/scripts --network postgres_net -it postgres psql -h db -U postgres
Password for user postgres: 
psql (17.4 (Debian 17.4-1.pgdg120+2))
Type "help" for help.

postgres=# create database awesome_portal_db_go
postgres-# ;
CREATE DATABASE
postgres=# 

Tu upgrade
/Users/frederic.beuserie/go/bin/migrate -path db/migrations -database "postgres://postgres:example@127.0.0.1/awesome_portal_db_go?sslmode=disable" up





Tools used:
- https://github.com/golang-migrate/migrate
- https://github.com/dal-go ??

