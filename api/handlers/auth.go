package handlers

import (
	"net/http"

	"ytmusic_api/middleware"
	"ytmusic_api/models"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

// AuthHandler holds dependencies for authentication endpoints.
type AuthHandler struct {
	Store  *ytmusic.SessionStore
	Client *ytmusic.Client
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(store *ytmusic.SessionStore, client *ytmusic.Client) *AuthHandler {
	return &AuthHandler{
		Store:  store,
		Client: client,
	}
}

// Login godoc
// @Summary      Authenticate with YouTube Music
// @Description  Submit your browser Cookie header string to create an authenticated session.
// @Description  Copy it from a POST request to music.youtube.com in browser DevTools (Network tab).
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body models.LoginRequest true "Browser cookies"
// @Success      200 {object} models.LoginResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid request body: " + err.Error(),
			Code:  http.StatusBadRequest,
		})
		return
	}

	session, err := h.Store.CreateSession(req.Cookies)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: err.Error(),
			Code:  http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:     session.Token,
		ExpiresAt: session.ExpiresAt.Format("2006-01-02T15:04:05Z"),
	})
}

// Logout godoc
// @Summary      Logout and invalidate session
// @Description  Removes the current session, requiring re-authentication.
// @Tags         auth
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.MessageResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /auth/logout [delete]
func (h *AuthHandler) Logout(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	h.Store.DeleteSession(session.Token)
	c.JSON(http.StatusOK, models.MessageResponse{Message: "logged out"})
}

// Status godoc
// @Summary      Check authentication status
// @Description  Returns whether the current session is valid and the associated account.
// @Tags         auth
// @Produce      json
// @Security     ApiKeyAuth
// @Success      200 {object} models.AuthStatusResponse
// @Failure      401 {object} models.ErrorResponse
// @Router       /auth/status [get]
func (h *AuthHandler) Status(c *gin.Context) {
	session := middleware.GetSession(c)
	if session == nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "not authenticated",
			Code:  http.StatusUnauthorized,
		})
		return
	}

	c.JSON(http.StatusOK, models.AuthStatusResponse{
		Authenticated: true,
	})
}
