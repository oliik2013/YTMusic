package middleware

import (
	"log/slog"
	"net"
	"net/http"

	"ytmusic_api/config"
	"ytmusic_api/models"
	"ytmusic_api/ytmusic"

	"github.com/gin-gonic/gin"
)

const (
	SessionTokenHeader  = "X-Session-Token"
	SessionKey          = "session"
	NewSessionTokenHeader = "X-New-Session-Token"
)

func AuthRequired(store *ytmusic.SessionStore, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader(SessionTokenHeader)

		if token == "" && isLocalhost(c) && cfg.Auth.Cookies != "" {
			session := store.GetAnySession()
			if session != nil {
				c.Set(SessionKey, session)
				c.Next()
				return
			}
			session, err := store.CreateSession(cfg.Auth.Cookies)
			if err == nil {
				c.Set(SessionKey, session)
				c.Next()
				return
			}
		}

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
				Error:     "missing X-Session-Token header",
				Code:      http.StatusUnauthorized,
				ErrorCode: models.ErrorCodeSessionExpired,
			})
			return
		}

		session := store.GetSession(token)
		if session == nil {
			if cfg.Auth.Cookies != "" && isLocalhost(c) {
				if err := cfg.Reload(); err != nil {
					slog.Debug("failed to reload config during auto-refresh", "error", err)
				}
				
				currentHash := cfg.CookiesHash()
				if store.CookiesChanged(currentHash) {
					newSession, err := store.RefreshFromConfig(cfg.Auth.Cookies)
					if err == nil {
						c.Header(NewSessionTokenHeader, newSession.Token)
						c.Set(SessionKey, newSession)
						slog.Info("auto-refreshed session from updated config", "token", newSession.Token)
						c.Next()
						return
					}
					slog.Warn("auto-refresh failed with updated config", "error", err)
				}
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:     "invalid or expired session token",
				Code:      http.StatusUnauthorized,
				ErrorCode: models.ErrorCodeSessionExpired,
			})
			return
		}

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
