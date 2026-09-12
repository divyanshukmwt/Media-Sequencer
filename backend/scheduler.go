package main

import "time"

func CurrentPlaylistItem(playlist []MediaItem, cycleStart time.Time, now time.Time) (item MediaItem, elapsedInItem time.Duration, ok bool) {
	if len(playlist) == 0 {
		return MediaItem{}, 0, false
	}

	totalDuration := 0
	for _, m := range playlist {
		if m.DurationSeconds > 0 {
			totalDuration += m.DurationSeconds
		}
	}
	if totalDuration <= 0 {
		return MediaItem{}, 0, false
	}

	elapsed := now.Sub(cycleStart)
	if elapsed < 0 {
		elapsed = 0
	}

	cycleSeconds := int64(CycleDuration.Seconds())
	elapsedSeconds := int64(elapsed.Seconds())
	cyclePos := elapsedSeconds % cycleSeconds

	playlistPos := cyclePos % int64(totalDuration)

	var acc int64
	for _, m := range playlist {
		d := int64(m.DurationSeconds)
		if d <= 0 {
			continue
		}
		acc += d
		if playlistPos < acc {
			return m, time.Duration(playlistPos-(acc-d)) * time.Second, true
		}
	}

	last := playlist[len(playlist)-1]
	return last, 0, true
}

func ResolveDisplay(win Window, sync SyncState, now time.Time) (item MediaItem, syncing bool) {
	if sync.Active && sync.Media != nil {
		expiry := sync.StartedAt.Add(time.Duration(sync.DurationSeconds) * time.Second)
		if now.Before(expiry) {
			return *sync.Media, true
		}
	}
	resolved, _, ok := CurrentPlaylistItem(win.Playlist, win.CycleStartAt, now)
	if !ok {
		return MediaItem{Type: MediaBlank, DurationSeconds: DefaultBlankDuration}, false
	}
	return resolved, false
}

func SyncRemainingSeconds(sync SyncState, now time.Time) int {
	if !sync.Active || sync.Media == nil {
		return 0
	}
	expiry := sync.StartedAt.Add(time.Duration(sync.DurationSeconds) * time.Second)
	remaining := expiry.Sub(now)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Seconds() + 0.999)
}
