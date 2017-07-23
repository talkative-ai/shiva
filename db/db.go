package db

import (
	"github.com/artificial-universe-maker/shiva/models"
	"github.com/go-gorp/gorp"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Required for sqlx postgres connections
)

// Shiva is the PostgreSQL connection instance
var Shiva *sqlx.DB
var DBMap *gorp.DbMap

// InitializeDB will setup the DB connection
func InitializeDB() error {
	var err error
	Shiva, err = sqlx.Connect("postgres", "user=postgres dbname=postgres host=postgres sslmode=disable")
	if err != nil {
		return err
	}

	DBMap = &gorp.DbMap{Db: Shiva.DB, Dialect: gorp.PostgresDialect{}}

	DBMap.AddTableWithName(models.AumProject{}, "projects")
	DBMap.AddTableWithName(models.AumZone{}, "zones")
	DBMap.AddTableWithName(models.AumActor{}, "actors")
	DBMap.AddTableWithName(models.AumDialog{}, "dialogs")
	DBMap.AddTableWithName(models.AumNote{}, "notes")

	return nil
}
