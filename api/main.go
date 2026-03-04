package main

import (
	"fmt"
	"log/slog"
	"os"
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
	// Setup structured logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Check ffmpeg availability
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		logger.Warn("ffmpeg not found in PATH. Audio playback requires ffmpeg", "hint", "Install ffmpeg and ensure it is available in your PATH")
	}

	// Initialise core components
	sessionStore := ytmusic.NewSessionStore()
	ytClient := ytmusic.NewClient()

	audioPlayer, err := player.NewPlayer()
	if err != nil {
		logger.Error("failed to initialise audio player", "error", err)
		os.Exit(1)
	}

	// If cookies are pre-seeded in config, auto-login
	if cfg.Auth.Cookies != "" {
		session, err := sessionStore.CreateSession(cfg.Auth.Cookies)
		if err != nil {
			logger.Warn("pre-seeded cookies are invalid", "error", err)
		} else {
			logger.Info("auto-logged in with pre-seeded cookies", "token", session.Token)
		}
	}

	// Create handlers
	authHandler := handlers.NewAuthHandler(sessionStore, ytClient, cfg.Auth.Cookies)
	playerHandler := handlers.NewPlayerHandler(audioPlayer, ytClient)
	queueHandler := handlers.NewQueueHandler(audioPlayer, ytClient)
	playlistHandler := handlers.NewPlaylistHandler(audioPlayer, ytClient)
	searchHandler := handlers.NewSearchHandler(ytClient)

	// Setup Gin router
	r := gin.Default()

	// --- Public routes (no auth required) ---
	r.POST("/auth/login", authHandler.Login)
	r.GET("/auth/status", authHandler.Status)

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// --- Protected routes (auth required) ---
	auth := r.Group("/")
	auth.Use(middleware.AuthRequired(sessionStore, cfg.Auth.Cookies))
	{
		// Auth
		auth.GET("/user", authHandler.UserInfo)
		auth.DELETE("/auth/logout", authHandler.Logout)

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
		auth.POST("/playlists/:id/cache", playlistHandler.CachePlaylist)

		// Search
		auth.GET("/search", searchHandler.Search)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logger.Info("YouTube Music API starting", "addr", fmt.Sprintf("http://%s", addr))
	logger.Info("Swagger UI available", "addr", fmt.Sprintf("http://%s/swagger/index.html", addr))
	if err := r.Run(addr); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
