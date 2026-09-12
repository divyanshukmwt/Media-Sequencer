package main

import (
	"context"
	"log"
	"time"
)

func SeedIfEmpty(ctx context.Context, store Store) error {
	count, err := store.CountWindows(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Printf("seed: %d window(s) already present, skipping seed", count)
		return nil
	}

	now := time.Now().UTC()

	window1 := Window{
		WindowID: "window-1",
		Name:     "Window 1",
		Playlist: []MediaItem{
			{ID: "m1", Type: MediaImage, URL: "https://picsum.photos/id/1015/1280/720", Title: "M1 - Mountain River", DurationSeconds: 8},
			{ID: "m2", Type: MediaVideo, URL: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", Title: "M2 - Big Buck Bunny (sample)", DurationSeconds: 15},
			{ID: "m3", Type: MediaBlank, Title: "M3 - Blank", DurationSeconds: 5},
		},
		CycleStartAt: now,
	}

	window2 := Window{
		WindowID: "window-2",
		Name:     "Window 2",
		Playlist: []MediaItem{
			{ID: "m4", Type: MediaImage, URL: "https://picsum.photos/id/1039/1280/720", Title: "M4 - Forest Lake", DurationSeconds: 10},
			{ID: "m5", Type: MediaVideo, URL: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/ForBiggerBlazes.mp4", Title: "M5 - Sample Clip", DurationSeconds: 12},
		},
		CycleStartAt: now,
	}

	for _, w := range []Window{window1, window2} {
		if err := store.CreateWindow(ctx, w); err != nil {
			return err
		}
	}
	log.Println("seed: inserted window-1 (M1,M2,M3) and window-2 (M4,M5)")
	return nil
}
