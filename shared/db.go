package shared

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

var DB *sqlx.DB

func Connect() *sqlx.DB {
	settings := LoadConfig()
	db, err := sqlx.Connect("pgx", settings.DatabaseURL)
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}
	DB = db
	return db
}
