package player

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// CacheManager handles background downloading of audio tracks via yt-dlp.
type CacheManager struct {
	cacheDir   string
	mu         sync.Mutex
	inProgress map[string]bool
	queue      chan string
}

var DefaultCacheManager *CacheManager

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("failed to get user home directory for cache", "error", err)
		return
	}
	cacheDir := filepath.Join(home, ".ytmusic", "cache")
	DefaultCacheManager = NewCacheManager(cacheDir)
}

// NewCacheManager initializes a new CacheManager and starts background workers.
func NewCacheManager(cacheDir string) *CacheManager {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		slog.Error("failed to create cache directory", "error", err)
	}

	cm := &CacheManager{
		cacheDir:   cacheDir,
		inProgress: make(map[string]bool),
		// buffered channel to hold download requests
		queue: make(chan string, 10000),
	}

	// start 3 concurrent download workers
	for i := 0; i < 3; i++ {
		go cm.worker()
	}

	return cm
}

// QueueDownload adds a video ID to the background download queue if it's not already cached.
func (cm *CacheManager) QueueDownload(videoID string) {
	if cm.IsCached(videoID) {
		return
	}

	cm.mu.Lock()
	if cm.inProgress[videoID] {
		cm.mu.Unlock()
		return
	}
	cm.inProgress[videoID] = true
	cm.mu.Unlock()

	// add to channel without blocking
	select {
	case cm.queue <- videoID:
	default:
		slog.Warn("cache queue is full, skipping", "videoID", videoID)
		cm.mu.Lock()
		delete(cm.inProgress, videoID)
		cm.mu.Unlock()
	}
}

// IsCached returns true if the audio file exists in the cache directory.
func (cm *CacheManager) IsCached(videoID string) bool {
	path := cm.GetCachedPath(videoID)
	_, err := os.Stat(path)
	return err == nil
}

// GetCachedPath returns the absolute path to the cached audio file.
func (cm *CacheManager) GetCachedPath(videoID string) string {
	return filepath.Join(cm.cacheDir, videoID+".m4a")
}

func (cm *CacheManager) worker() {
	for videoID := range cm.queue {
		slog.Info("Caching track via yt-dlp", "videoID", videoID)

		path := cm.GetCachedPath(videoID)

		// Use yt-dlp to download and convert to m4a
		cmd := exec.Command("yt-dlp",
			"--quiet", "--no-warnings",
			"-x", "--audio-format", "m4a",
			"-o", path,
			"https://music.youtube.com/watch?v="+videoID,
		)

		if err := cmd.Run(); err != nil {
			slog.Error("Failed to cache track", "videoID", videoID, "error", err)
			// clean up any partial file
			_ = os.Remove(path)
		} else {
			slog.Info("Successfully cached track", "videoID", videoID)
		}

		cm.mu.Lock()
		delete(cm.inProgress, videoID)
		cm.mu.Unlock()
	}
}
