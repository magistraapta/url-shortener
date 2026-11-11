package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"url-shortener/controller"
	"url-shortener/database"
	"url-shortener/repositories"
	"url-shortener/services"

	"github.com/joho/godotenv"
)

func main() {

	// Local dev convenience: load .env if present. Missing file is fine —
	// in production the environment is set by the platform.
	_ = godotenv.Load()

	database := database.NewDatabase()
	if err := database.Connect(context.Background()); err != nil {
		log.Fatal(err)
	}
	db := database.DB


	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}

	repo := repositories.NewUrlRepository(db)
	svc := services.NewUrlService(repo)
	ctrl := controller.NewUrlController(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	ctrl.RegisterRoutes(mux)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
