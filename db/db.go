package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var Conn *sql.DB

func Init() {
	// Hardcoded for now — you can externalize later if needed
	dbUser := "root"
	dbPass := "123456"
	dbHost := "127.0.0.1"
	dbPort := "3306"
	dbName := "batch"

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPass, dbHost, dbPort, dbName)

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
