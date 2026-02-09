package main

import (
	"log"

	"buono-tax-invoice/internal/config"
	"buono-tax-invoice/internal/database"
	"buono-tax-invoice/internal/handlers"
	"buono-tax-invoice/internal/repository"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// ===== 1. Load Config =====
	cfg := config.LoadConfig()

	// ===== 2. Connect Database =====
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	// ===== 3. Run Migrations =====
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("❌ Migration failed: %v", err)
	}

	// ===== 4. Setup Repository & Handler =====
	customerRepo := repository.NewCustomerRepository(db)
	customerHandler := handlers.NewCustomerHandler(customerRepo)

	// ===== 5. Setup Gin Router =====
	router := gin.Default()

	// CORS - อนุญาตให้ Frontend เรียกได้
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	// ===== 6. Serve Frontend Static Files =====
	router.Static("/static", "./static")
	router.StaticFile("/", "./static/index.html")

	// ===== 7. API Routes =====
	api := router.Group("/api")
	{
		customer := api.Group("/customer")
		{
			customer.GET("/search", customerHandler.Search) // ค้นหาข้อมูล
			customer.POST("", customerHandler.Save)         // บันทึก/แก้ไขข้อมูล
		}
	}

	// ===== 8. Start Server =====
	port := ":" + cfg.ServerPort
	log.Printf("🚀 Buono Tax Invoice Server started on port %s", cfg.ServerPort)
	log.Printf("📄 Frontend: http://localhost%s", port)
	log.Printf("📡 API: http://localhost%s/api/customer/search?q=xxx", port)

	// Bind to 0.0.0.0 for container/cloud deployment
	if err := router.Run("0.0.0.0" + port); err != nil {
		log.Fatalf("❌ Server failed to start: %v", err)
	}
}
