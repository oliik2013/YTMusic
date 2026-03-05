package player

import (
	"fmt"
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
		return nil, nil
	}

	drpc, err := discordrpc.New(clientID)
	if err != nil {
		return nil, fmt.Errorf("creating Discord RPC client: %w", err)
	}

	return &DiscordPresence{
		client: drpc,
	}, nil
}

func (d *DiscordPresence) UpdatePresence(track *models.Track, isPlaying bool, positionMs int64) {
	if d == nil || d.client == nil {
		return
	}

	if !isPlaying {
		d.Clear()
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

	_ = d.client.SetActivity(activity)
}

func (d *DiscordPresence) Clear() {
	if d == nil || d.client == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	emptyActivity := discordrpc.Activity{}
	_ = d.client.SetActivity(emptyActivity)
}

func (d *DiscordPresence) Close() {
	if d == nil || d.client == nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	_ = d.client.Socket.Close()
}
