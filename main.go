// main func
package main

import (
	"context"
	"log"
	"os"

	"url-shortener/controller"
	"url-shortener/database"
	"url-shortener/repositories"
	"url-shortener/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Local dev convenience: load .env if present. Missing file is fine —
	// in production the environment is set by the platform.
	_ = godotenv.Load()

	db, err := database.ConnectDatabase(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	repo := repositories.NewURLRepository(db)
	svc := services.NewUrlService(repo, baseURL)
	ctrl := controller.NewURLController(svc)

	router := gin.Default()
	router.GET("/healthz", func(c *gin.Context) { c.Status(200) })
	ctrl.RegisterRoutes(router)

	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
