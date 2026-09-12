package main

import "time"

const (
	MediaImage = "image"
	MediaVideo = "video"
	MediaBlank = "blank"
)

const (
	DefaultImageDuration = 5
	DefaultVideoDuration = 10
	DefaultBlankDuration = 3
)

const CycleDuration = 5 * time.Hour

type MediaItem struct {
	ID              string `bson:"id" json:"id"`
	Type            string `bson:"type" json:"type"`
	URL             string `bson:"url,omitempty" json:"url,omitempty"`
	Title           string `bson:"title,omitempty" json:"title,omitempty"`
	DurationSeconds int    `bson:"duration_seconds" json:"duration_seconds"`
}

type Window struct {
	WindowID     string      `bson:"_id" json:"window_id"`
	Name         string      `bson:"name" json:"name"`
	Playlist     []MediaItem `bson:"playlist" json:"playlist"`
	CycleStartAt time.Time   `bson:"cycle_start_at" json:"cycle_start_at"`
	CreatedAt    time.Time   `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time   `bson:"updated_at" json:"updated_at"`
}

type SyncState struct {
	ID              string     `bson:"_id" json:"-"`
	Active          bool       `bson:"active" json:"active"`
	Media           *MediaItem `bson:"media,omitempty" json:"media,omitempty"`
	StartedAt       time.Time  `bson:"started_at,omitempty" json:"started_at,omitempty"`
	DurationSeconds int        `bson:"duration_seconds,omitempty" json:"duration_seconds,omitempty"`
}

const SyncStateDocID = "current"
