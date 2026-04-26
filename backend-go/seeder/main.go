package main

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 🔥 DI SINI letaknya
	dsn := "root:@tcp(127.0.0.1:3306)/product_db?parseTime=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = SeedProducts(db, 100000)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Seeding selesai 🚀")
}