package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ytmusic_api/models"

	"github.com/gin-gonic/gin"
)

type LyricsHandler struct {
	httpClient *http.Client
}

func NewLyricsHandler() *LyricsHandler {
	return &LyricsHandler{
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type lrcLibResult struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	TrackName      string  `json:"trackName"`
	ArtistName     string  `json:"artistName"`
	AlbumName      string  `json:"albumName"`
	Duration       float64 `json:"duration"`
	Instrumental   bool    `json:"instrumental"`
	PlainLyrics    string  `json:"plainLyrics"`
	SyncedLyrics   string  `json:"syncedLyrics"`
}

// GetLyrics godoc
// @Summary      Get lyrics for a track
// @Description  Searches LrcLib for lyrics by track name and artist. Returns synced lyrics if available, otherwise plain lyrics.
// @Tags         lyrics
// @Produce      json
// @Param        track_name  query string true "Track name"
// @Param        artist_name query string true "Artist name"
// @Param        album_name  query string false "Album name (optional)"
// @Success      200 {object} models.LyricsResponse
// @Failure      400 {object} models.ErrorResponse
// @Failure      404 {object} models.ErrorResponse
// @Failure      500 {object} models.ErrorResponse
// @Router       /lyrics [get]
func (h *LyricsHandler) GetLyrics(c *gin.Context) {
	trackName := strings.TrimSpace(c.Query("track_name"))
	artistName := strings.TrimSpace(c.Query("artist_name"))
	albumName := strings.TrimSpace(c.Query("album_name"))

	if trackName == "" || artistName == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "track_name and artist_name are required",
			Code:  http.StatusBadRequest,
		})
		return
	}

	cleanedTitle := cleanTitle(trackName)
	cleanedArtist := cleanArtist(artistName)

	var results []lrcLibResult

	results = h.searchLrcLib(cleanedTitle, cleanedArtist, albumName)
	if len(results) == 0 {
		results = h.searchLrcLib(cleanedTitle, cleanedArtist, "")
	}
	if len(results) == 0 {
		results = h.searchLrcLibQuery(cleanedArtist + " " + cleanedTitle)
	}
	if len(results) == 0 {
		results = h.searchLrcLibQuery(cleanedTitle)
	}
	if len(results) == 0 && (cleanedTitle != trackName || cleanedArtist != artistName) {
		results = h.searchLrcLib(trackName, artistName, albumName)
	}

	for _, r := range results {
		if r.SyncedLyrics != "" || r.PlainLyrics != "" {
			response := models.LyricsResponse{
				ID:           r.ID,
				TrackName:    r.TrackName,
				ArtistName:   r.ArtistName,
				AlbumName:    r.AlbumName,
				Duration:     r.Duration,
				Instrumental: r.Instrumental,
				PlainLyrics:  r.PlainLyrics,
				SyncedLyrics: r.SyncedLyrics,
				Source:       "lrclib",
			}

			if r.SyncedLyrics != "" {
				response.ParsedLyrics = parseLRC(r.SyncedLyrics)
			} else if r.PlainLyrics != "" {
				response.ParsedLyrics = parsePlainLyrics(r.PlainLyrics)
			}

			c.JSON(http.StatusOK, response)
			return
		}
	}

	c.JSON(http.StatusNotFound, models.ErrorResponse{
		Error: "no lyrics found for this track",
		Code:  http.StatusNotFound,
	})
}

func (h *LyricsHandler) searchLrcLib(trackName, artistName, albumName string) []lrcLibResult {
	baseURL := "https://lrclib.net/api/search"
	params := url.Values{}
	params.Set("track_name", trackName)
	params.Set("artist_name", artistName)
	if albumName != "" {
		params.Set("album_name", albumName)
	}

	return h.doSearch(baseURL + "?" + params.Encode())
}

func (h *LyricsHandler) searchLrcLibQuery(query string) []lrcLibResult {
	baseURL := "https://lrclib.net/api/search"
	params := url.Values{}
	params.Set("q", query)
	return h.doSearch(baseURL + "?" + params.Encode())
}

func (h *LyricsHandler) doSearch(urlStr string) []lrcLibResult {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		slog.Debug("failed to create request", "error", err)
		return nil
	}
	req.Header.Set("User-Agent", "YTMusic-API/1.0 (https://github.com/user/ytmusic)")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		slog.Debug("failed to fetch from lrclib", "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("lrclib returned non-200", "status", resp.StatusCode)
		return nil
	}

	var results []lrcLibResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		slog.Debug("failed to decode lrclib response", "error", err)
		return nil
	}

	return results
}

func parseLRC(lrc string) []models.LyricsLine {
	lines := strings.Split(lrc, "\n")
	var parsed []models.LyricsLine

	lrcRegex := regexp.MustCompile(`\[(\d{2}):(\d{2})\.(\d{2,3})\](.*)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		matches := lrcRegex.FindStringSubmatch(line)
		if len(matches) == 5 {
			minutes, _ := strconv.Atoi(matches[1])
			seconds, _ := strconv.Atoi(matches[2])
			msStr := matches[3]
			var ms int
			if len(msStr) == 2 {
				ms, _ = strconv.Atoi(msStr)
				ms *= 10
			} else {
				ms, _ = strconv.Atoi(msStr)
			}

			timeMs := int64(minutes*60*1000 + seconds*1000 + ms)
			text := strings.TrimSpace(matches[4])

			if text != "" {
				parsed = append(parsed, models.LyricsLine{
					TimeMs: timeMs,
					Text:   text,
				})
			}
		}
	}

	return parsed
}

func parsePlainLyrics(lyrics string) []models.LyricsLine {
	lines := strings.Split(lyrics, "\n")
	var parsed []models.LyricsLine

	for i, line := range lines {
		text := strings.TrimSpace(line)
		if text != "" {
			parsed = append(parsed, models.LyricsLine{
				TimeMs: int64(i * 5000),
				Text:   text,
			})
		}
	}

	return parsed
}

func cleanTitle(title string) string {
	title = strings.TrimSpace(title)

	featRegex := regexp.MustCompile(`(?i)\s*[\(\[]?\s*(feat\.?|featuring|ft\.?)\s+[^\)\]]*[\)\]]?`)
	title = featRegex.ReplaceAllString(title, "")

	title = strings.TrimSpace(title)
	return title
}

func cleanArtist(artist string) string {
	artist = strings.TrimSpace(artist)

	featRegex := regexp.MustCompile(`(?i)\s*[,/&]\s*.*`)
	artist = featRegex.ReplaceAllString(artist, "")

	artist = strings.TrimSpace(artist)
	return artist
}
