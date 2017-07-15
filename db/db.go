package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Required for sqlx postgres connections
)

//
var ShivaDB *sqlx.DB

var initializeSchema = `
	CREATE TABLE IF NOT EXISTS Person (
		first_name text,
		last_name text,
		email text
	);

	CREATE TABLE IF NOT EXISTS Place (
		country text,
		city text NULL,
		telcode integer
	)
`

// InitializeDB will setup the DB connection
func InitializeDB() error {
	ShivaDB, err := sqlx.Connect("postgres", "user=postgres dbname=postgres host=postgres sslmode=disable")
	if err != nil {
		return err
	}

	ShivaDB.MustExec(initializeSchema)

	return nil
}
