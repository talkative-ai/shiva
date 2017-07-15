package db

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Required for sqlx postgres connections
)

// ShivaDB is the PostgreSQL connection instance
var ShivaDB *sqlx.DB

var initializeSchema = `
	CREATE TABLE IF NOT EXISTS projects (
		id SERIAL PRIMARY KEY,
		title text,
		owner_id int,
		start_zone int,
		created_at timestamp DEFAULT current_timestamp
	);
	CREATE TABLE IF NOT EXISTS actors (
		id SERIAL PRIMARY KEY,
		title text,
		created_at timestamp DEFAULT current_timestamp
	);
	CREATE TABLE IF NOT EXISTS dialogs (
		id SERIAL PRIMARY KEY,
		title text,
		created_at timestamp DEFAULT current_timestamp
	);
	CREATE TABLE IF NOT EXISTS zones (
		id SERIAL PRIMARY KEY,
		title text,
		created_at timestamp DEFAULT current_timestamp
	);
	CREATE TABLE IF NOT EXISTS notes (
		id SERIAL PRIMARY KEY,
		title text,
		content text,
		created_at timestamp DEFAULT current_timestamp
	);

	CREATE TABLE IF NOT EXISTS project_actors (
		project_id int NOT NULL references projects(id),
		actor_id int NOT NULL references actors(id)
	);
	CREATE TABLE IF NOT EXISTS project_dialogs (
		project_id int NOT NULL references projects(id),
		dialog_id int NOT NULL references dialogs(id)
	);
	CREATE TABLE IF NOT EXISTS project_zones (
		project_id int NOT NULL references projects(id),
		zone_id int NOT NULL references zones(id)
	);
	CREATE TABLE IF NOT EXISTS project_notes (
		project_id int NOT NULL references projects(id),
		note_id int NOT NULL references notes(id)
	);
`

// InitializeDB will setup the DB connection
func InitializeDB() error {
	var err error
	ShivaDB, err = sqlx.Connect("postgres", "user=postgres dbname=postgres host=postgres sslmode=disable")
	if err != nil {
		return err
	}

	ShivaDB.MustExec(initializeSchema)

	return nil
}
