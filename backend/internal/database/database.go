package database

import (
"database/sql"

_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect(databaseURL string) (*sql.DB, error) {
return sql.Open("pgx", databaseURL)
}
