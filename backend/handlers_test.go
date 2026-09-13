package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestRouter(t *testing.T) (http.Handler, *fakeStore) {
	t.Helper()
	store := newFakeStore()
	ctx := context.Background()
	if err := SeedIfEmpty(ctx, store); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	api := NewAPI(store)
	return NewRouter(api, "*"), store
}

func doRequest(t *testing.T, h http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestHealth(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "GET", "/health", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestListWindows_SeededData(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "GET", "/api/windows", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var windows []Window
	if err := json.Unmarshal(rec.Body.Bytes(), &windows); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(windows) != 2 {
		t.Fatalf("expected 2 seeded windows, got %d", len(windows))
	}
}

func TestGetState_ReturnsCurrentItemPerWindow(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "GET", "/api/state", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var state stateResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(state.Windows) != 2 {
		t.Fatalf("expected 2 windows in state, got %d", len(state.Windows))
	}
	for _, w := range state.Windows {
		if w.CurrentItem.ID == "" {
			t.Errorf("window %s has no current item resolved", w.WindowID)
		}
	}
}

func TestAddMedia_AppendsToPlaylist(t *testing.T) {
	h, store := setupTestRouter(t)
	reqBody := addMediaRequest{Type: MediaImage, URL: "https://example.com/x.jpg", Title: "New Image", DurationSeconds: 7}
	rec := doRequest(t, h, "POST", "/api/windows/window-1/media", reqBody)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	win, _ := store.GetWindow(context.Background(), "window-1")
	if len(win.Playlist) != 4 {
		t.Fatalf("expected 4 items after add (3 seeded + 1), got %d", len(win.Playlist))
	}
	last := win.Playlist[len(win.Playlist)-1]
	if last.Title != "New Image" || last.DurationSeconds != 7 {
		t.Errorf("unexpected last item: %+v", last)
	}
}

func TestAddMedia_RejectsInvalidType(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "POST", "/api/windows/window-1/media", map[string]any{"type": "audio", "url": "x"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid type, got %d", rec.Code)
	}
}

func TestAddMedia_RejectsMissingURLForImage(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "POST", "/api/windows/window-1/media", map[string]any{"type": "image"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing url, got %d", rec.Code)
	}
}

func TestAddMedia_UnknownWindow404s(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "POST", "/api/windows/does-not-exist/media", addMediaRequest{Type: MediaBlank})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTriggerSync_ThenStateReflectsSyncOnAllWindows(t *testing.T) {
	h, _ := setupTestRouter(t)

	rec := doRequest(t, h, "POST", "/api/sync", triggerSyncRequest{MediaID: "m2", DurationSeconds: 5})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	stateRec := doRequest(t, h, "GET", "/api/state", nil)
	var state stateResponse
	if err := json.Unmarshal(stateRec.Body.Bytes(), &state); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !state.Sync.Active {
		t.Fatal("expected sync to be active")
	}
	for _, w := range state.Windows {
		if !w.Syncing {
			t.Errorf("window %s should report syncing=true", w.WindowID)
		}
		if w.CurrentItem.ID != "m2" {
			t.Errorf("window %s should show m2 during sync, got %s", w.WindowID, w.CurrentItem.ID)
		}
	}
}

func TestTriggerSync_UnknownMediaID404s(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "POST", "/api/sync", triggerSyncRequest{MediaID: "does-not-exist"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRemoveMedia_RemovesJustThatItem(t *testing.T) {
	h, store := setupTestRouter(t)
	rec := doRequest(t, h, "DELETE", "/api/windows/window-1/media/m2", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	win, _ := store.GetWindow(context.Background(), "window-1")
	if len(win.Playlist) != 2 {
		t.Fatalf("expected 2 items left after removing m2, got %d", len(win.Playlist))
	}
	for _, m := range win.Playlist {
		if m.ID == "m2" {
			t.Fatal("m2 should have been removed")
		}
	}
}

func TestRemoveMedia_UnknownWindow404s(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "DELETE", "/api/windows/does-not-exist/media/m2", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestRenameWindow(t *testing.T) {
	h, store := setupTestRouter(t)
	rec := doRequest(t, h, "PATCH", "/api/windows/window-1", renameWindowRequest{Name: "Lobby Display"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	win, _ := store.GetWindow(context.Background(), "window-1")
	if win.Name != "Lobby Display" {
		t.Fatalf("expected renamed window, got %q", win.Name)
	}
}

func TestRenameWindow_RejectsEmptyName(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "PATCH", "/api/windows/window-1", renameWindowRequest{Name: "  "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestRenameWindow_UnknownWindow404s(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "PATCH", "/api/windows/does-not-exist", renameWindowRequest{Name: "X"})
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestDeleteWindow(t *testing.T) {
	h, store := setupTestRouter(t)
	rec := doRequest(t, h, "DELETE", "/api/windows/window-2", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, err := store.GetWindow(context.Background(), "window-2"); err != ErrNotFound {
		t.Fatalf("expected window-2 to be gone, got err=%v", err)
	}
	windows, _ := store.ListWindows(context.Background())
	if len(windows) != 1 {
		t.Fatalf("expected 1 window left, got %d", len(windows))
	}
}

func TestDeleteWindow_UnknownWindow404s(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "DELETE", "/api/windows/does-not-exist", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestCreateWindow(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "POST", "/api/windows", createWindowRequest{Name: "Window 3"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var win Window
	if err := json.Unmarshal(rec.Body.Bytes(), &win); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if win.Name != "Window 3" || len(win.Playlist) != 0 {
		t.Errorf("unexpected created window: %+v", win)
	}
}

func TestCreateWindow_RejectsEmptyName(t *testing.T) {
	h, _ := setupTestRouter(t)
	rec := doRequest(t, h, "POST", "/api/windows", createWindowRequest{Name: "  "})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}