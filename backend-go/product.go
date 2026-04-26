package handlers

import (
    "database/sql"
    "net/http"
    "product-crud-api/config"
    "product-crud-api/models"
    "strconv"

    "github.com/gin-gonic/gin"
)

func GetProducts(c *gin.Context) {
    rows, err := config.DB.Query("SELECT id, kode_item, nama_produk, qty_produk, harga_produk FROM products ORDER BY id DESC")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var products []models.Product
    for rows.Next() {
        var product models.Product
        err := rows.Scan(&product.ID, &product.KodeItem, &product.NamaProduk, &product.QtyProduk, &product.HargaProduk)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        products = append(products, product)
    }

    c.JSON(http.StatusOK, products)
}

func CreateProduct(c *gin.Context) {
    var product models.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := config.DB.Exec("INSERT INTO products (kode_item, nama_produk, qty_produk, harga_produk) VALUES (?, ?, ?, ?)",
        product.KodeItem, product.NamaProduk, product.QtyProduk, product.HargaProduk)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    id, _ := result.LastInsertId()
    product.ID = int(id)
    c.JSON(http.StatusCreated, product)
}

func UpdateProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    var product models.Product
    if err := c.ShouldBindJSON(&product); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    _, err = config.DB.Exec("UPDATE products SET kode_item = ?, nama_produk = ?, qty_produk = ?, harga_produk = ? WHERE id = ?",
        product.KodeItem, product.NamaProduk, product.QtyProduk, product.HargaProduk, id)
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    product.ID = id
    c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    _, err = config.DB.Exec("DELETE FROM products WHERE id = ?", id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}