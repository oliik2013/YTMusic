package handlers

import (
	"net/http"

	"ytmusic_api/middleware"
	"ytmusic_api/models"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

type BrowseHandler struct {
	Client *ytmusic.Client
}

func NewBrowseHandler(client *ytmusic.Client) *BrowseHandler {
	return &BrowseHandler{
		Client: client,
	}
}

// GetArtist godoc
// @Summary      Get artist details
// @Description  Returns artist information including top tracks and albums.
// @Tags         browse
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Artist browse ID (starts with UC)"
// @Success      200 {object} models.ArtistDetail
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /artists/{id} [get]
func (h *BrowseHandler) GetArtist(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	browseID := c.Param("id")
	artist, err := h.Client.GetArtist(session, browseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch artist: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, artist)
}

// GetAlbum godoc
// @Summary      Get album details
// @Description  Returns album information including all tracks.
// @Tags         browse
// @Produce      json
// @Security     ApiKeyAuth
// @Param        id path string true "Album browse ID (starts with MPREb)"
// @Success      200 {object} models.AlbumDetail
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /albums/{id} [get]
func (h *BrowseHandler) GetAlbum(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	browseID := c.Param("id")
	album, err := h.Client.GetAlbum(session, browseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch album: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, album)
}
