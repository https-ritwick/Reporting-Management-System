package db

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
)

var Conn *sql.DB

func Init() {

	fmt.Println("✅ Using env from:", filepath.Join(".", ".env"))

	// Read from actual .env
	dbUser := "root"
	dbPass := "123456"
	dbHost := "127.0.0.1"
	dbPort := "3306"
	dbName := "batch"

	// Print to confirm
	fmt.Println("DB_USER:", dbUser)
	fmt.Println("DB_PASS:", dbPass)
	fmt.Println("DB_HOST:", dbHost)
	fmt.Println("DB_PORT:", dbPort)
	fmt.Println("DB_NAME:", dbName)

	// Build DSN correctly
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	fmt.Println("DSN:", dsn)
	var err error
	Conn, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}

	if err = Conn.Ping(); err != nil {
		log.Fatalf("DB ping error: %v", err)
	}

	log.Println("✅ Connected to MySQL DB")
}
