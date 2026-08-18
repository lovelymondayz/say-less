package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/lovelymondayz/say-less/backend/internal/handlers"
	"github.com/lovelymondayz/say-less/backend/internal/models"
	"github.com/lovelymondayz/say-less/backend/internal/spotify"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5437/sayless?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	if err := models.Migrate(pool); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Connected to database")

	spotifyClient := spotify.NewClient()

	container := &models.Container{
		DB:      pool,
		Spotify: spotifyClient,
	}

	h := handlers.New(container)

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	r.GET("/api/health", h.Health)
	r.POST("/api/generate", h.Generate)
	r.POST("/api/share", h.SaveShare)
	r.GET("/api/share/:id", h.GetShare)
	r.GET("/api/spotify/login", h.SpotifyLogin)
	r.GET("/api/spotify/callback", h.SpotifyCallback)
	r.POST("/api/spotify/playlist", h.CreatePlaylist)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
