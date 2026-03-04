package ytmusic

import (
	"crypto/sha1"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	ytMusicOrigin = "https://music.youtube.com"
)

// Session represents an authenticated user session.
type Session struct {
	Token     string
	Cookies   string
	SAPISID   string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// SessionStore manages active sessions in memory.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session // keyed by token
}

// NewSessionStore creates a new empty session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
	}
}

// CreateSession parses the cookie string, extracts SAPISID, and stores a new session.
// Returns the session or an error if SAPISID is not found.
func (s *SessionStore) CreateSession(cookies string) (*Session, error) {
	sapisid := extractCookieValue(cookies, "SAPISID")
	if sapisid == "" {
		// Also try __Secure-3PAPISID which is the same value
		sapisid = extractCookieValue(cookies, "__Secure-3PAPISID")
	}
	if sapisid == "" {
		return nil, fmt.Errorf("SAPISID cookie not found in provided cookies; ensure you copied all request headers from music.youtube.com")
	}

	token := uuid.New().String()
	session := &Session{
		Token:     token,
		Cookies:   cookies,
		SAPISID:   sapisid,
		ExpiresAt: time.Now().Add(24 * 365 * time.Hour), // ~1 year, matching cookie lifespan
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.sessions[token] = session
	s.mu.Unlock()

	return session, nil
}

// GetSession retrieves a session by token. Returns nil if not found or expired.
func (s *SessionStore) GetSession(token string) *Session {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[token]
	if !ok {
		return nil
	}
	if time.Now().After(session.ExpiresAt) {
		return nil
	}
	return session
}

// DeleteSession removes a session by token.
func (s *SessionStore) DeleteSession(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// GetAuthorizationHeader computes the SAPISIDHASH authorization header value.
// Format: SAPISIDHASH <timestamp>_<sha1(timestamp + " " + SAPISID + " " + origin)>
func GetAuthorizationHeader(sapisid string) string {
	ts := time.Now().Unix()
	input := fmt.Sprintf("%d %s %s", ts, sapisid, ytMusicOrigin)
	hash := sha1.Sum([]byte(input))
	return fmt.Sprintf("SAPISIDHASH %d_%x", ts, hash)
}

// extractCookieValue parses a raw cookie header string and returns the value of the named cookie.
func extractCookieValue(cookieStr, name string) string {
	// Cookie header format: "name1=value1; name2=value2; ..."
	pairs := strings.Split(cookieStr, ";")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if idx := strings.Index(pair, "="); idx > 0 {
			key := strings.TrimSpace(pair[:idx])
			val := strings.TrimSpace(pair[idx+1:])
			if key == name {
				return val
			}
		}
	}
	return ""
}
