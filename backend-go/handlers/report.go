package handlers

import (
    "database/sql"
    "net/http"
    "product-crud-api/config"
    "time"

    "github.com/gin-gonic/gin"
)

func GetProductsByDateRange(c *gin.Context) {
    startDate := c.Query("start_date")
    endDate := c.Query("end_date")
    
    var rows *sql.Rows
    var err error
    
    if startDate != "" && endDate != "" {
        query := "SELECT id, kode_item, nama_produk, qty_produk, harga_produk, created_at FROM products WHERE DATE(created_at) BETWEEN ? AND ? ORDER BY created_at DESC"
        rows, err = config.DB.Query(query, startDate, endDate)
    } else {
        query := "SELECT id, kode_item, nama_produk, qty_produk, harga_produk, created_at FROM products ORDER BY created_at DESC"
        rows, err = config.DB.Query(query)
    }
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()
    
    var products []map[string]interface{}
    for rows.Next() {
        var id int
        var kodeItem, namaProduk string
        var qtyProduk int
        var hargaProduk float64
        var createdAt time.Time
        
        err := rows.Scan(&id, &kodeItem, &namaProduk, &qtyProduk, &hargaProduk, &createdAt)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        
        products = append(products, map[string]interface{}{
            "id":           id,
            "kode_item":    kodeItem,
            "nama_produk":  namaProduk,
            "qty_produk":   qtyProduk,
            "harga_produk": hargaProduk,
            "created_at":   createdAt.Format("2006-01-02"),
        })
    }
    
    c.JSON(http.StatusOK, products)
}
