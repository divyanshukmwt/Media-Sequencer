package main

import (
	"context"
	"sync"
)

type fakeStore struct {
	mu      sync.Mutex
	windows map[string]Window
	sync    SyncState
}

func newFakeStore() *fakeStore {
	return &fakeStore{windows: map[string]Window{}}
}

func (f *fakeStore) ListWindows(ctx context.Context) ([]Window, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := []Window{}
	for _, w := range f.windows {
		out = append(out, w)
	}
	return out, nil
}

func (f *fakeStore) GetWindow(ctx context.Context, id string) (*Window, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w, ok := f.windows[id]
	if !ok {
		return nil, ErrNotFound
	}
	return &w, nil
}

func (f *fakeStore) CreateWindow(ctx context.Context, w Window) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if w.Playlist == nil {
		w.Playlist = []MediaItem{}
	}
	f.windows[w.WindowID] = w
	return nil
}

func (f *fakeStore) AddMedia(ctx context.Context, windowID string, item MediaItem) (*Window, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w, ok := f.windows[windowID]
	if !ok {
		return nil, ErrNotFound
	}
	w.Playlist = append(w.Playlist, item)
	f.windows[windowID] = w
	return &w, nil
}

func (f *fakeStore) FindMediaByID(ctx context.Context, mediaID string) (*MediaItem, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, w := range f.windows {
		for _, m := range w.Playlist {
			if m.ID == mediaID {
				mCopy := m
				return &mCopy, nil
			}
		}
	}
	return nil, ErrNotFound
}

func (f *fakeStore) GetSyncState(ctx context.Context) (SyncState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sync, nil
}

func (f *fakeStore) SetSyncState(ctx context.Context, s SyncState) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sync = s
	return nil
}

func (f *fakeStore) CountWindows(ctx context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return int64(len(f.windows)), nil
}
