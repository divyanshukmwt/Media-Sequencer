package main

import (
	"testing"
	"time"
)

func sampleWindow1() []MediaItem {
	return []MediaItem{
		{ID: "m1", Type: MediaImage, DurationSeconds: 10},
		{ID: "m2", Type: MediaVideo, DurationSeconds: 20},
		{ID: "m3", Type: MediaBlank, DurationSeconds: 10},
	}
}

func TestCurrentPlaylistItem_FirstItemAtStart(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	item, elapsed, ok := CurrentPlaylistItem(sampleWindow1(), start, start)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if item.ID != "m1" {
		t.Errorf("expected m1 at t=0, got %s", item.ID)
	}
	if elapsed != 0 {
		t.Errorf("expected 0 elapsed, got %v", elapsed)
	}
}

func TestCurrentPlaylistItem_MidFirstItem(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(5 * time.Second)
	item, elapsed, _ := CurrentPlaylistItem(sampleWindow1(), start, now)
	if item.ID != "m1" || elapsed != 5*time.Second {
		t.Errorf("expected m1 at 5s in, got %s at %v", item.ID, elapsed)
	}
}

func TestCurrentPlaylistItem_SecondItem(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(13 * time.Second)
	item, elapsed, _ := CurrentPlaylistItem(sampleWindow1(), start, now)
	if item.ID != "m2" || elapsed != 3*time.Second {
		t.Errorf("expected m2 at 3s in, got %s at %v", item.ID, elapsed)
	}
}

func TestCurrentPlaylistItem_LoopsBackToStart(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(45 * time.Second)
	item, elapsed, _ := CurrentPlaylistItem(sampleWindow1(), start, now)
	if item.ID != "m1" || elapsed != 5*time.Second {
		t.Errorf("expected loop back to m1 at 5s, got %s at %v", item.ID, elapsed)
	}
}

func TestCurrentPlaylistItem_ManyLoopsLater(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(2 * time.Hour)
	item, elapsed, _ := CurrentPlaylistItem(sampleWindow1(), start, now)
	if item.ID != "m1" || elapsed != 0 {
		t.Errorf("expected exact loop boundary -> m1 at 0s, got %s at %v", item.ID, elapsed)
	}
}

func TestCurrentPlaylistItem_WrapsAtFiveHourCycleBoundary(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(5 * time.Hour)
	item, elapsed, _ := CurrentPlaylistItem(sampleWindow1(), start, now)
	if item.ID != "m1" || elapsed != 0 {
		t.Errorf("expected cycle wrap -> m1 at 0s, got %s at %v", item.ID, elapsed)
	}
}

func TestCurrentPlaylistItem_JustBeforeFiveHourBoundary(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(5*time.Hour - time.Second)
	item, elapsed, _ := CurrentPlaylistItem(sampleWindow1(), start, now)
	if item.ID != "m3" || elapsed != 9*time.Second {
		t.Errorf("expected m3 at 9s in, got %s at %v", item.ID, elapsed)
	}
}

func TestCurrentPlaylistItem_EmptyPlaylist(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, _, ok := CurrentPlaylistItem(nil, start, start)
	if ok {
		t.Error("expected ok=false for empty playlist")
	}
}

func TestCurrentPlaylistItem_ZeroDurationItemsSkipped(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	playlist := []MediaItem{
		{ID: "bad", Type: MediaImage, DurationSeconds: 0},
		{ID: "good", Type: MediaImage, DurationSeconds: 5},
	}
	item, _, ok := CurrentPlaylistItem(playlist, start, start)
	if !ok || item.ID != "good" {
		t.Errorf("expected zero-duration item to be skipped, got %v ok=%v", item, ok)
	}
}

func TestResolveDisplay_NoSyncUsesOwnSchedule(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	win := Window{WindowID: "w1", Playlist: sampleWindow1(), CycleStartAt: start}
	sync := SyncState{Active: false}
	item, syncing := ResolveDisplay(win, sync, start)
	if syncing {
		t.Error("expected not syncing")
	}
	if item.ID != "m1" {
		t.Errorf("expected own m1, got %s", item.ID)
	}
}

func TestResolveDisplay_ActiveSyncOverridesEveryWindow(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	win1 := Window{WindowID: "w1", Playlist: sampleWindow1(), CycleStartAt: start}
	win2 := Window{WindowID: "w2", Playlist: []MediaItem{{ID: "m4", DurationSeconds: 8}, {ID: "m5", DurationSeconds: 12}}, CycleStartAt: start}
	syncedMedia := MediaItem{ID: "m2", Type: MediaVideo, DurationSeconds: 20}
	sync := SyncState{Active: true, Media: &syncedMedia, StartedAt: start, DurationSeconds: 10}

	now := start.Add(3 * time.Second)
	item1, syncing1 := ResolveDisplay(win1, sync, now)
	item2, syncing2 := ResolveDisplay(win2, sync, now)

	if !syncing1 || !syncing2 {
		t.Error("expected both windows to report syncing=true")
	}
	if item1.ID != "m2" || item2.ID != "m2" {
		t.Errorf("expected both windows to show m2, got %s and %s", item1.ID, item2.ID)
	}
}

func TestResolveDisplay_ResumesOwnPlaylistAfterSyncExpires(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	win := Window{WindowID: "w1", Playlist: sampleWindow1(), CycleStartAt: start}
	syncedMedia := MediaItem{ID: "m2", Type: MediaVideo, DurationSeconds: 20}
	sync := SyncState{Active: true, Media: &syncedMedia, StartedAt: start, DurationSeconds: 10}

	now := start.Add(11 * time.Second)
	item, syncing := ResolveDisplay(win, sync, now)
	if syncing {
		t.Error("expected sync to have expired")
	}
	if item.ID != "m2" {
		t.Errorf("expected window's own schedule to resume at m2, got %s", item.ID)
	}
	if len(win.Playlist) != 3 || win.Playlist[0].ID != "m1" {
		t.Error("original playlist was mutated by sync -- it must never be")
	}
}

func TestSyncRemainingSeconds(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	media := MediaItem{ID: "m2"}
	sync := SyncState{Active: true, Media: &media, StartedAt: start, DurationSeconds: 10}

	if got := SyncRemainingSeconds(sync, start.Add(2*time.Second)); got != 8 {
		t.Errorf("expected 8s remaining, got %d", got)
	}
	if got := SyncRemainingSeconds(sync, start.Add(15*time.Second)); got != 0 {
		t.Errorf("expected 0s remaining after expiry, got %d", got)
	}
	if got := SyncRemainingSeconds(SyncState{Active: false}, start); got != 0 {
		t.Errorf("expected 0s when sync inactive, got %d", got)
	}
}
