package handlers

import (
	"net/http"
	"strconv"

	"ytmusic_api/middleware"
	"ytmusic_api/models"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

// SearchHandler holds dependencies for the search endpoint.
type SearchHandler struct {
	Client *ytmusic.Client
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(client *ytmusic.Client) *SearchHandler {
	return &SearchHandler{
		Client: client,
	}
}

// Search godoc
// @Summary      Search YouTube Music
// @Description  Searches YouTube Music for songs, albums, artists, or playlists.
// @Description  Use the filter parameter to narrow results to a specific type.
// @Tags         search
// @Produce      json
// @Security     ApiKeyAuth
// @Param        q      query string true  "Search query"
// @Param        filter query string false "Filter type: songs, albums, artists, playlists, videos"
// @Param        limit  query int    false "Max results to return (default 20)"
// @Success      200 {object} models.SearchResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /search [get]
func (h *SearchHandler) Search(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "query parameter 'q' is required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	filter := c.Query("filter")

	results, err := h.Client.Search(session, query, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "search failed: " + err.Error(),
			Code:  http.StatusInternalServerError,
		})
		return
	}

	// Apply limit if specified
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit < len(results) {
			results = results[:limit]
		}
	}

	c.JSON(http.StatusOK, models.SearchResponse{
		Results: results,
		Query:   query,
	})
}
