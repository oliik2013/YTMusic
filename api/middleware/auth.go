package middleware

import (
	"net/http"

	"ytmusic_api/models"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

const (
	// SessionTokenHeader is the HTTP header clients must send to authenticate.
	SessionTokenHeader = "X-Session-Token"

	// SessionKey is the gin context key where the session is stored after auth.
	SessionKey = "session"
)

// AuthRequired returns a Gin middleware that validates the X-Session-Token header
// against the given SessionStore. If valid, the session is stored in the context.
func AuthRequired(store *ytmusic.SessionStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(SessionTokenHeader)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing X-Session-Token header",
				Code:  http.StatusUnauthorized,
			})
			return
		}

		session := store.GetSession(token)
		if session == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired session token",
				Code:  http.StatusUnauthorized,
			})
			return
		}

		// Store session in context for handlers to use
		c.Set(SessionKey, session)
		c.Next()
	}
}

// GetSession retrieves the authenticated session from the Gin context.
func GetSession(c *gin.Context) *ytmusic.Session {
	val, exists := c.Get(SessionKey)
	if !exists {
		return nil
	}
	session, ok := val.(*ytmusic.Session)
	if !ok {
		return nil
	}
	return session
}
