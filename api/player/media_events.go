package player

import (
	"log/slog"
)

// MediaEvent represents a hardware/media key event.
// Actions supported: "pause", "next", "previous", "play" (resume if paused).
type MediaEvent struct {
	Action string
}

// Internal channel for media events. Platform-specific code can push events here.
var mediaEventCh = make(chan MediaEvent, 32)

// SendMediaEvent injects a media event into the dispatcher pipeline.
func SendMediaEvent(e MediaEvent) {
	mediaEventCh <- e
}

// StartMediaEventDispatcher starts a goroutine that translates media events into
// playback actions on the provided Player instance.
func StartMediaEventDispatcher(p *Player) {
	for e := range mediaEventCh {
		switch e.Action {
		case "pause":
			p.Pause()
		case "next":
			if _, err := p.Next(); err != nil {
				slog.Error("media event: next failed", "error", err)
			}
		case "previous":
			if _, err := p.Previous(); err != nil {
				slog.Error("media event: previous failed", "error", err)
			}
		case "play":
			// Resume if paused
			if p.isPaused {
				p.Pause()
			}
		default:
			slog.Debug("media event: unrecognized action", "action", e.Action)
		}
	}
}
