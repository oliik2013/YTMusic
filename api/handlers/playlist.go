package handlers

import (
	"net/http"

	"ytmusic_api/middleware"
	"ytmusic_api/models"
	"ytmusic_api/player"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

// PlaylistHandler holds dependencies for playlist endpoints.
type PlaylistHandler struct {
	Player *player.Player
	Client *ytmusic.Client
}

// NewPlaylistHandler creates a new PlaylistHandler.
func NewPlaylistHandler(p *player.Player, client *ytmusic.Client) *PlaylistHandler {
	return &PlaylistHandler{
		Player: p,
		Client: client,
	}
}

// ListPlaylists godoc
// @Summary      List user's playlists
// @Description  Returns all playlists from the authenticated user's YouTube Music library.
// @Tags         playlists
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.PlaylistListResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /playlists [get]
func (h *PlaylistHandler) ListPlaylists(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	playlists, err := h.Client.GetLibraryPlaylists(session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch playlists: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, models.PlaylistListResponse{Playlists: playlists})
}

// GetPlaylist godoc
// @Summary      Get playlist details
// @Description  Returns the metadata and full track listing for a playlist by ID.
// @Tags         playlists
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Playlist ID"
// @Success      200 {object} models.PlaylistDetailResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /playlists/{id} [get]
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	playlistID := c.Param("id")
	detail, err := h.Client.GetPlaylist(session, playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch playlist: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// PlayPlaylist godoc
// @Summary      Play a playlist
// @Description  Loads all tracks from the playlist into the queue and begins playback from the first track.
// @Tags         playlists
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Playlist ID"
// @Success      200 {object} models.PlayerState
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /playlists/{id}/play [post]
func (h *PlaylistHandler) PlayPlaylist(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	playlistID := c.Param("id")
	detail, err := h.Client.GetPlaylist(session, playlistID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch playlist: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	if len(detail.Tracks) == 0 {
		c.JSON(http.StatusOK, models.ErrorResponse{
			Error: "playlist is empty",
			Code:  http.StatusOK,
		})
		return
	}

	// Replace queue with playlist tracks
	h.Player.Queue.ReplaceAll(detail.Tracks)

	// Start playing the first track
	first := &detail.Tracks[0]
	if err := h.Player.PlayTrack(first); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to start playback: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, h.Player.State())
}
