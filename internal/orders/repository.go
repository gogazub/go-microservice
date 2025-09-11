package orders

import (
	"database/sql"
)

type PostgresOrderRepository struct {
	db *sql.DB
}
