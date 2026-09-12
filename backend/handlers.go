package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

type API struct {
	store Store
}

func NewAPI(store Store) *API {
	return &API{store: store}
}


func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func newID(prefix string) string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	return prefix + "-" + hex.EncodeToString(b)
}


func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}


type windowStateDTO struct {
	WindowID          string      `json:"window_id"`
	Name              string      `json:"name"`
	Playlist          []MediaItem `json:"playlist"`
	CycleStartAt      time.Time   `json:"cycle_start_at"`
	CurrentItem       MediaItem   `json:"current_item"`
	ElapsedInItemSecs int         `json:"elapsed_in_item_seconds"`
	Syncing           bool        `json:"syncing"`
}

type syncStateDTO struct {
	Active            bool       `json:"active"`
	Media             *MediaItem `json:"media,omitempty"`
	DurationSeconds   int        `json:"duration_seconds,omitempty"`
	RemainingSeconds  int        `json:"remaining_seconds"`
}

type stateResponse struct {
	ServerTime time.Time         `json:"server_time"`
	Windows    []windowStateDTO  `json:"windows"`
	Sync       syncStateDTO      `json:"sync"`
}

func (a *API) getState(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	now := time.Now().UTC()

	windows, err := a.store.ListWindows(ctx)
	if err != nil {
		log.Printf("getState: list windows: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load windows")
		return
	}
	sync, err := a.store.GetSyncState(ctx)
	if err != nil {
		log.Printf("getState: get sync: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load sync state")
		return
	}

	resp := stateResponse{ServerTime: now, Windows: []windowStateDTO{}}
	for _, win := range windows {
		current, syncing := ResolveDisplay(win, sync, now)
		_, elapsed, ok := CurrentPlaylistItem(win.Playlist, win.CycleStartAt, now)
		elapsedSecs := 0
		if ok && !syncing {
			elapsedSecs = int(elapsed.Seconds())
		}
		resp.Windows = append(resp.Windows, windowStateDTO{
			WindowID:          win.WindowID,
			Name:              win.Name,
			Playlist:          win.Playlist,
			CycleStartAt:      win.CycleStartAt,
			CurrentItem:       current,
			ElapsedInItemSecs: elapsedSecs,
			Syncing:           syncing,
		})
	}
	resp.Sync = syncStateDTO{
		Active:           sync.Active,
		Media:            sync.Media,
		DurationSeconds:  sync.DurationSeconds,
		RemainingSeconds: SyncRemainingSeconds(sync, now),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (a *API) listWindows(w http.ResponseWriter, r *http.Request) {
	windows, err := a.store.ListWindows(r.Context())
	if err != nil {
		log.Printf("listWindows: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load windows")
		return
	}
	writeJSON(w, http.StatusOK, windows)
}


type createWindowRequest struct {
	Name string `json:"name"`
}

func (a *API) createWindow(w http.ResponseWriter, r *http.Request) {
	var req createWindowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	win := Window{
		WindowID: newID("window"),
		Name:     req.Name,
		Playlist: []MediaItem{},
	}
	if err := a.store.CreateWindow(r.Context(), win); err != nil {
		log.Printf("createWindow: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create window")
		return
	}
	writeJSON(w, http.StatusCreated, win)
}


func (a *API) getWindow(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	win, err := a.store.GetWindow(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "window not found")
		return
	}
	if err != nil {
		log.Printf("getWindow: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load window")
		return
	}
	writeJSON(w, http.StatusOK, win)
}


type addMediaRequest struct {
	Type            string `json:"type"`
	URL             string `json:"url"`
	Title           string `json:"title"`
	DurationSeconds int    `json:"duration_seconds"`
}

func defaultDurationFor(mediaType string) int {
	switch mediaType {
	case MediaImage:
		return DefaultImageDuration
	case MediaVideo:
		return DefaultVideoDuration
	case MediaBlank:
		return DefaultBlankDuration
	}
	return DefaultImageDuration
}

func (a *API) addMedia(w http.ResponseWriter, r *http.Request) {
	windowID := r.PathValue("id")
	var req addMediaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.Type = strings.ToLower(strings.TrimSpace(req.Type))
	if req.Type != MediaImage && req.Type != MediaVideo && req.Type != MediaBlank {
		writeError(w, http.StatusBadRequest, "type must be one of: image, video, blank")
		return
	}
	if (req.Type == MediaImage || req.Type == MediaVideo) && strings.TrimSpace(req.URL) == "" {
		writeError(w, http.StatusBadRequest, "url is required for image/video media")
		return
	}
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = defaultDurationFor(req.Type)
	}

	item := MediaItem{
		ID:              newID("media"),
		Type:            req.Type,
		URL:             req.URL,
		Title:           req.Title,
		DurationSeconds: req.DurationSeconds,
	}

	win, err := a.store.AddMedia(r.Context(), windowID, item)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "window not found")
		return
	}
	if err != nil {
		log.Printf("addMedia: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to add media")
		return
	}
	writeJSON(w, http.StatusCreated, win)
}


func (a *API) getSync(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	sync, err := a.store.GetSyncState(r.Context())
	if err != nil {
		log.Printf("getSync: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to load sync state")
		return
	}
	writeJSON(w, http.StatusOK, syncStateDTO{
		Active:           sync.Active,
		Media:            sync.Media,
		DurationSeconds:  sync.DurationSeconds,
		RemainingSeconds: SyncRemainingSeconds(sync, now),
	})
}


const DefaultSyncDurationSeconds = 10

type triggerSyncRequest struct {
	MediaID         string `json:"media_id"`
	DurationSeconds int    `json:"duration_seconds"`
}

func (a *API) triggerSync(w http.ResponseWriter, r *http.Request) {
	var req triggerSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	req.MediaID = strings.TrimSpace(req.MediaID)
	if req.MediaID == "" {
		writeError(w, http.StatusBadRequest, "media_id is required")
		return
	}
	if req.DurationSeconds <= 0 {
		req.DurationSeconds = DefaultSyncDurationSeconds
	}

	ctx := r.Context()
	media, err := a.store.FindMediaByID(ctx, req.MediaID)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "media_id not found in any window's playlist")
		return
	}
	if err != nil {
		log.Printf("triggerSync: find media: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to look up media")
		return
	}

	now := time.Now().UTC()
	state := SyncState{
		Active:          true,
		Media:           media,
		StartedAt:       now,
		DurationSeconds: req.DurationSeconds,
	}
	if err := a.store.SetSyncState(ctx, state); err != nil {
		log.Printf("triggerSync: set sync: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to trigger sync")
		return
	}
	writeJSON(w, http.StatusOK, syncStateDTO{
		Active:           true,
		Media:            media,
		DurationSeconds:  req.DurationSeconds,
		RemainingSeconds: req.DurationSeconds,
	})
}


func NewRouter(a *API, frontendOrigin string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", a.health)
	mux.HandleFunc("GET /api/state", a.getState)
	mux.HandleFunc("GET /api/windows", a.listWindows)
	mux.HandleFunc("POST /api/windows", a.createWindow)
	mux.HandleFunc("GET /api/windows/{id}", a.getWindow)
	mux.HandleFunc("POST /api/windows/{id}/media", a.addMedia)
	mux.HandleFunc("GET /api/sync", a.getSync)
	mux.HandleFunc("POST /api/sync", a.triggerSync)

	return withCORS(withLogging(mux), frontendOrigin)
}

func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%v)", r.Method, r.URL.Path, time.Since(start))
	})
}
