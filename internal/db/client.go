package db

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

type DatabaseConnection struct {
	Client *sql.DB
}

func NewDatabaseConn() *DatabaseConnection {
	dbc := &DatabaseConnection{nil}
	host := os.Getenv("DATABASE_HOST")

	if host == "" {
		log.Fatalln(`Could not retrieve env "DATABASE_HOST"!`)
	}

	port := os.Getenv("DATABASE_PORT")

	if port == "" {
		log.Fatalln(`Could not retrieve env "DATABASE_PORT"!`)
	}

	database := os.Getenv("DATABASE_NAME")

	if database == "" {
		log.Fatalln(`Could not retrieve env "DATABASE_NAME"!`)
	}

	username := os.Getenv("DATABASE_USER")

	if username == "" {
		log.Fatalln(`Could not retrieve env "DATABASE_USER"!`)
	}

	password := os.Getenv("DATABASE_PASSWORD")

	if password == "" {
		log.Fatalln(`Could not retrieve env "DATABASE_PASSWORD"!`)
	}

	dburi := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, database)

	db, err := sql.Open("mysql", dburi)

	if err != nil {
		log.Fatalf(`Failed to establish connection to database: "%s"\n`, err.Error())
	}

	dbc.Client = db

	log.Println("Database connection established, client has been set...")

	return dbc
}
