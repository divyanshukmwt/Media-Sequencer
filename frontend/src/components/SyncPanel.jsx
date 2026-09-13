import { useEffect, useState } from "react";
import { triggerSync } from "../api";

export default function SyncPanel({ windows, sync, onSynced }) {
  const [selectedId, setSelectedId] = useState(null);
  const [duration, setDuration] = useState(10);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  const allMedia = windows.flatMap((w) =>
    w.playlist.map((m) => ({ ...m, windowName: w.name }))
  );
  useEffect(() => {
    if (selectedId && !allMedia.some((m) => m.id === selectedId)) {
      setSelectedId(null);
    }
  }, [windows]);

  async function handleTake() {
    if (!selectedId) {
      setError("Pick a media item first.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const result = await triggerSync(selectedId, Number(duration) || 10);
      onSynced(result);
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="rack">
      <div className="rack-header">
        <span className="rack-title">SYNC / TAKE</span>
        <span className={`rack-status ${sync.active ? "active" : "idle"}`}>
          {sync.active
            ? `SYNCED to ${sync.media?.title || sync.media?.id} — ${sync.remaining_seconds}s remaining`
            : "idle — every window on its own schedule"}
        </span>
      </div>

      {allMedia.length === 0 ? (
        <p className="form-error" style={{ margin: 0 }}>
          Add media to a window first — sync picks from what's already in a playlist.
        </p>
      ) : (
        <div className="media-picker">
          {allMedia.map((m) => (
            <button
              key={m.id}
              className={`media-chip ${selectedId === m.id ? "selected" : ""}`}
              onClick={() => setSelectedId(m.id)}
              title={`from ${m.windowName}`}
            >
              {m.title || m.id}
            </button>
          ))}
        </div>
      )}

      <div className="rack-controls">
        <label>
          Duration
          <input
            className="field"
            type="number"
            min="1"
            value={duration}
            onChange={(e) => setDuration(e.target.value)}
          />
          sec
        </label>
        <button className="btn btn-take" onClick={handleTake} disabled={submitting || allMedia.length === 0}>
          {submitting ? "Taking…" : "TAKE TO ALL MONITORS"}
        </button>
      </div>
      {error && <div className="form-error" style={{ marginTop: 8 }}>{error}</div>}
    </div>
  );
}