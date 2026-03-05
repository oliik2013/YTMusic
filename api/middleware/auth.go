package middleware

import (
	"net"
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
// For localhost requests with pre-seeded cookies in config, auth is skipped entirely.
// For localhost requests with a Cookie header, auth is skipped if a session can be created.
func AuthRequired(store *ytmusic.SessionStore, preSeededCookies string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(SessionTokenHeader)

		// For localhost requests with pre-seeded cookies, skip auth entirely
		if token == "" && isLocalhost(c) && preSeededCookies != "" {
			session := store.GetAnySession()
			if session != nil {
				c.Set(SessionKey, session)
				c.Next()
				return
			}
			// Try to create a session from pre-seeded cookies
			session, err := store.CreateSession(preSeededCookies)
			if err == nil {
				c.Set(SessionKey, session)
				c.Next()
				return
			}
		}

		// For localhost requests with a Cookie header, create session from cookie
		if token == "" && isLocalhost(c) {
			if cookie := c.GetHeader("Cookie"); cookie != "" {
				session, err := store.CreateSession(cookie)
				if err == nil {
					c.Set(SessionKey, session)
					c.Next()
					return
				}
			}
		}

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

func isLocalhost(c *gin.Context) bool {
	ip := c.ClientIP()
	return ip == "127.0.0.1" || ip == "::1" || ip == "::ffff:127.0.0.1" || net.ParseIP(ip).IsLoopback()
}
