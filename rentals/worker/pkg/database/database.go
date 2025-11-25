package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "postgresql"
	port     = 5432
	user     = "okteto"
	password = "okteto"
	dbname   = "rentals"
)

func Open() (*sql.DB, error) {
	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	return sql.Open("postgres", psqlconn)
}

func LoadData(db *sql.DB) error {
	// Create rentals table
	dropRentalsStmt := `DROP TABLE IF EXISTS rentals`
	if _, err := db.Exec(dropRentalsStmt); err != nil {
		return err
	}

	createRentalsStmt := `CREATE TABLE IF NOT EXISTS rentals (id VARCHAR(255) NOT NULL UNIQUE, price VARCHAR(255) NOT NULL)`
	if _, err := db.Exec(createRentalsStmt); err != nil {
		return err
	}

	return nil
}

func Ping(db *sql.DB) {
	fmt.Println("Waiting for postgresql...")
	for {
		if err := db.Ping(); err == nil {
			fmt.Println("Postgresql connected!")
			return
		}

		time.Sleep(1 * time.Second)
	}
}

// createOrUpdateRental creates or updates a rental entry in the database
func CreateOrUpdateRental(db *sql.DB, rentalID string, rentalPrice string) error {
	fmt.Println("Received internal request to create/update rental...")

	insertDynStmt := `insert into "rentals"("id", "price") values($1, $2) on conflict(id) do update set price = $2`
	if _, err := db.Exec(insertDynStmt, rentalID, rentalPrice); err != nil {
		return err
	}

	return nil
}

// deleteRental deletes a rental entry from the database
func DeleteRental(db *sql.DB, rentalID string) error {
	fmt.Printf("Received internal request to delete rental: ID=%s\n", rentalID)
	deleteStmt := `DELETE FROM rentals WHERE id = $1`
	if _, err := db.Exec(deleteStmt, rentalID); err != nil {
		return err
	}

	return nil
}
