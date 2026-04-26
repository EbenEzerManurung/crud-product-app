package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
    _ "github.com/go-sql-driver/mysql"
    "github.com/joho/godotenv"
)

type Product struct {
    ID          int       `json:"id"`
    KodeItem    string    `json:"kode_item"`
    NamaProduk  string    `json:"nama_produk"`
    QtyProduk   int       `json:"qty_produk"`
    HargaProduk float64   `json:"harga_produk"`
    CreatedAt   time.Time `json:"created_at"`
}

var DB *sql.DB

func initDB() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found, using defaults")
    }

    // Get database config from environment
    dbUser := getEnv("DB_USER", "root")
    dbPassword := getEnv("DB_PASSWORD", "")
    dbHost := getEnv("DB_HOST", "localhost")
    dbPort := getEnv("DB_PORT", "3306")
    dbName := getEnv("DB_NAME", "product_db")

    // Create DSN
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        dbUser, dbPassword, dbHost, dbPort, dbName)

    // Open connection
    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal("Error opening database: ", err)
    }

    // Test connection
    if err = DB.Ping(); err != nil {
        log.Fatal("Error connecting to database: ", err)
    }

    fmt.Println("✅ Successfully connected to MySQL database")

    // Create table if not exists
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS products (
        id INT AUTO_INCREMENT PRIMARY KEY,
        kode_item VARCHAR(50) NOT NULL UNIQUE,
        nama_produk VARCHAR(100) NOT NULL,
        qty_produk INT NOT NULL DEFAULT 0,
        harga_produk DECIMAL(15,2) NOT NULL DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
    ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
    `

    if _, err = DB.Exec(createTableSQL); err != nil {
        log.Fatal("Error creating table: ", err)
    }
    fmt.Println("✅ Products table ready")
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// Health check
func healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok", "message": "Server is running"})
}

// Get all products
func getProducts(c *gin.Context) {
    rows, err := DB.Query("SELECT id, kode_item, nama_produk, qty_produk, harga_produk, created_at FROM products ORDER BY id DESC")
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    defer rows.Close()

    var products []Product
    for rows.Next() {
        var p Product
        if err := rows.Scan(&p.ID, &p.KodeItem, &p.NamaProduk, &p.QtyProduk, &p.HargaProduk, &p.CreatedAt); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        products = append(products, p)
    }

    if products == nil {
        products = []Product{}
    }
    c.JSON(http.StatusOK, products)
}

// Get products by date range for report
func getProductsByDateRange(c *gin.Context) {
    startDate := c.Query("start_date")
    endDate := c.Query("end_date")
    
    var rows *sql.Rows
    var err error
    
    if startDate != "" && endDate != "" {
        query := "SELECT id, kode_item, nama_produk, qty_produk, harga_produk, created_at FROM products WHERE DATE(created_at) BETWEEN ? AND ? ORDER BY created_at DESC"
        rows, err = DB.Query(query, startDate, endDate)
    } else {
        query := "SELECT id, kode_item, nama_produk, qty_produk, harga_produk, created_at FROM products ORDER BY created_at DESC"
        rows, err = DB.Query(query)
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

// Create product
func createProduct(c *gin.Context) {
    var p Product
    if err := c.ShouldBindJSON(&p); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := DB.Exec("INSERT INTO products (kode_item, nama_produk, qty_produk, harga_produk) VALUES (?, ?, ?, ?)",
        p.KodeItem, p.NamaProduk, p.QtyProduk, p.HargaProduk)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    id, _ := result.LastInsertId()
    p.ID = int(id)
    p.CreatedAt = time.Now()
    c.JSON(http.StatusCreated, p)
}

// Update product
func updateProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    var p Product
    if err := c.ShouldBindJSON(&p); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := DB.Exec("UPDATE products SET kode_item=?, nama_produk=?, qty_produk=?, harga_produk=? WHERE id=?",
        p.KodeItem, p.NamaProduk, p.QtyProduk, p.HargaProduk, id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    p.ID = id
    c.JSON(http.StatusOK, p)
}

// Delete product
func deleteProduct(c *gin.Context) {
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
        return
    }

    result, err := DB.Exec("DELETE FROM products WHERE id=?", id)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func main() {
    // Initialize database
    initDB()
    defer DB.Close()

    // Create Gin router
    r := gin.Default()

    // CORS configuration
    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
    }))

    // Routes
    api := r.Group("/api")
    {
        api.GET("/products", getProducts)
        api.GET("/products/report", getProductsByDateRange)
        api.POST("/products", createProduct)
        api.PUT("/products/:id", updateProduct)
        api.DELETE("/products/:id", deleteProduct)
    }
    r.GET("/health", healthCheck)

    // Start server
    fmt.Println("🚀 Server starting on http://localhost:8080")
    fmt.Println("📊 API Endpoints:")
    fmt.Println("  - GET    /api/products")
    fmt.Println("  - GET    /api/products/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD")
    fmt.Println("  - POST   /api/products")
    fmt.Println("  - PUT    /api/products/:id")
    fmt.Println("  - DELETE /api/products/:id")
    fmt.Println("  - GET    /health")
    
    log.Fatal(r.Run(":8080"))
}
