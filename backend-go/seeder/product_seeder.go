package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"
)

func SeedProducts(db *sql.DB, total int) error {
	batchSize := 1000

	rand.Seed(time.Now().UnixNano())

	for i := 0; i < total/batchSize; i++ {
		tx, err := db.Begin()
		if err != nil {
			return err
		}

		query := "INSERT INTO products (kode_item, nama_produk, qty_produk, harga_produk) VALUES "
		values := []interface{}{}

		for j := 0; j < batchSize; j++ {
			query += "(?, ?, ?, ?),"

			index := i*batchSize + j

			kode := fmt.Sprintf("PRD%06d", index)
			nama := fmt.Sprintf("Produk-%d", rand.Intn(100000))
			qty := rand.Intn(100) + 1
			harga := rand.Intn(1000000) + 1000

			values = append(values, kode, nama, qty, harga)
		}

		// hapus koma terakhir
		query = query[:len(query)-1]

		_, err = tx.Exec(query, values...)
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		fmt.Printf("Batch %d selesai\n", i+1)
	}

	return nil
}