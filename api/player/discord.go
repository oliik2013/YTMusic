package player

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"ytmusic_api/models"

	"github.com/rikkuness/discord-rpc"
)

type DiscordPresence struct {
	client *discordrpc.Client
	mu     sync.Mutex
}

func NewDiscordPresence(clientID string) (*DiscordPresence, error) {
	if clientID == "" {
		slog.Debug("Discord Client ID not configured, skipping Discord RPC initialization")
		return nil, nil
	}

	drpc, err := discordrpc.New(clientID)
	if err != nil {
		return nil, fmt.Errorf("creating Discord RPC client: %w", err)
	}

	slog.Info("Discord RPC connected", "clientID", clientID)

	return &DiscordPresence{
		client: drpc,
	}, nil
}

func (d *DiscordPresence) UpdatePresence(track *models.Track, isPlaying bool, positionMs int64) {
	if d == nil || d.client == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if track == nil {
		return
	}

	activity := discordrpc.Activity{
		Type:    2, // Listening to
		Details: track.Title,
		State:   fmt.Sprintf("by %s", track.Artist),
		Assets: &discordrpc.Assets{
			LargeImage: track.ThumbnailURL,
			LargeText:  track.Title,
		},
	}

	if track.DurationMs > 0 {
		startTime := discordrpc.Epoch{Time: time.Now().Add(-time.Duration(positionMs) * time.Millisecond)}
		endTime := discordrpc.Epoch{Time: time.Now().Add(time.Duration(track.DurationMs-positionMs) * time.Millisecond)}
		activity.Timestamps = &discordrpc.Timestamps{
			Start: &startTime,
			End:   &endTime,
		}
	}

	if err := d.client.SetActivity(activity); err != nil {
		slog.Error("failed to set Discord activity", "error", err)
	}
}

func (d *DiscordPresence) Clear() {
	if d == nil || d.client == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	emptyActivity := discordrpc.Activity{}
	if err := d.client.SetActivity(emptyActivity); err != nil {
		slog.Error("failed to clear Discord activity", "error", err)
	}
}

func (d *DiscordPresence) Close() {
	if d == nil || d.client == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.client.Socket.Close(); err != nil {
		slog.Error("failed to close Discord RPC connection", "error", err)
	}
}
