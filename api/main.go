package main

import (
	"fmt"
	"log"
	"os/exec"

	"ytmusic_api/config"
	"ytmusic_api/handlers"
	"ytmusic_api/middleware"
	"ytmusic_api/player"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "ytmusic_api/docs"
)

// @title           YouTube Music API
// @version         1.0
// @description     A self-contained YouTube Music player backend that authenticates via browser cookies,
// @description     resolves audio streams through the Innertube internal API, and plays audio locally.
// @description     Designed as a modular music player daemon for TUIs, Raycast extensions, and other clients.

// @contact.name    YTMusic API
// @license.name    MIT

// @host            localhost:8080
// @BasePath        /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Session-Token

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Check ffmpeg availability
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		log.Println("WARNING: ffmpeg not found in PATH. Audio playback requires ffmpeg.")
		log.Println("         Install ffmpeg and ensure it is available in your PATH.")
	}

	// Initialise core components
	sessionStore := ytmusic.NewSessionStore()
	ytClient := ytmusic.NewClient()

	audioPlayer, err := player.NewPlayer()
	if err != nil {
		log.Fatalf("Failed to initialise audio player: %v", err)
	}

	// If cookies are pre-seeded in config, auto-login
	if cfg.Auth.Cookies != "" {
		session, err := sessionStore.CreateSession(cfg.Auth.Cookies)
		if err != nil {
			log.Printf("WARNING: Pre-seeded cookies are invalid: %v", err)
		} else {
			log.Printf("Auto-logged in with pre-seeded cookies (token: %s)", session.Token)
		}
	}

	// Create handlers
	authHandler := handlers.NewAuthHandler(sessionStore, ytClient)
	playerHandler := handlers.NewPlayerHandler(audioPlayer, ytClient)
	queueHandler := handlers.NewQueueHandler(audioPlayer, ytClient)
	playlistHandler := handlers.NewPlaylistHandler(audioPlayer, ytClient)
	searchHandler := handlers.NewSearchHandler(ytClient)

	// Setup Gin router
	r := gin.Default()

	// --- Public routes (no auth required) ---
	r.POST("/auth/login", authHandler.Login)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// --- Protected routes (auth required) ---
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(sessionStore))
	{
		// Auth
		auth.DELETE("/auth/logout", authHandler.Logout)
		auth.GET("/auth/status", authHandler.Status)

		// Player
		auth.POST("/player/play", playerHandler.Play)
		auth.POST("/player/pause", playerHandler.PauseToggle)
		auth.POST("/player/next", playerHandler.NextTrack)
		auth.POST("/player/previous", playerHandler.PreviousTrack)
		auth.POST("/player/stop", playerHandler.Stop)
		auth.POST("/player/volume", playerHandler.SetVolume)
		auth.GET("/player/state", playerHandler.GetState)

		// Queue
		auth.GET("/queue", queueHandler.GetQueue)
		auth.POST("/queue/add", queueHandler.AddToQueue)
		auth.DELETE("/queue", queueHandler.ClearQueue)
		auth.DELETE("/queue/:position", queueHandler.RemoveFromQueue)

		// Playlists
		auth.GET("/playlists", playlistHandler.ListPlaylists)
		auth.GET("/playlists/:id", playlistHandler.GetPlaylist)
		auth.POST("/playlists/:id/play", playlistHandler.PlayPlaylist)

		// Search
		auth.GET("/search", searchHandler.Search)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("YouTube Music API starting on http://%s", addr)
	log.Printf("Swagger UI available at http://%s/swagger/index.html", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
