import { useState } from "react";
import AddMediaForm from "./AddMediaForm";
import { formatTimecode } from "../../format";
import { removeMedia, renameWindow, deleteWindow } from "../../api";

export default function MonitorWindow({ win, onChanged }) {
  const [addingMedia, setAddingMedia] = useState(false);
  const [showPlaylist, setShowPlaylist] = useState(false);
  const [editingName, setEditingName] = useState(false);
  const [nameValue, setNameValue] = useState(win.name);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState(null);

  const current = win.current_item;
  const hasPlaylist = win.playlist && win.playlist.length > 0;
  const elapsed = win.elapsed_in_item_seconds || 0;
  const duration = current?.duration_seconds || 0;
  const progressPct = duration > 0 ? Math.min(100, (elapsed / duration) * 100) : 0;

  async function saveName() {
    const trimmed = nameValue.trim();
    setEditingName(false);
    if (!trimmed || trimmed === win.name) {
      setNameValue(win.name);
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await renameWindow(win.window_id, trimmed);
      onChanged();
    } catch (err) {
      setError(err.message);
      setNameValue(win.name);
    } finally {
      setBusy(false);
    }
  }

  async function handleDeleteWindow() {
    if (!window.confirm(`Delete "${win.name}" and its whole playlist? This can't be undone.`)) {
      return;
    }
    setBusy(true);
    setError(null);
    try {
      await deleteWindow(win.window_id);
      onChanged();
    } catch (err) {
      setError(err.message);
      setBusy(false);
    }
  }

  async function handleRemoveMedia(mediaId) {
    setBusy(true);
    setError(null);
    try {
      await removeMedia(win.window_id, mediaId);
      onChanged();
    } catch (err) {
      setError(err.message);
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className={`monitor ${win.syncing ? "is-synced" : ""}`}>
      <div className="monitor-label-plate">
        {editingName ? (
          <input
            className="field monitor-name-input"
            value={nameValue}
            autoFocus
            disabled={busy}
            onChange={(e) => setNameValue(e.target.value)}
            onBlur={saveName}
            onKeyDown={(e) => {
              if (e.key === "Enter") e.currentTarget.blur();
              if (e.key === "Escape") {
                setNameValue(win.name);
                setEditingName(false);
              }
            }}
          />
        ) : (
          <button
            className="monitor-name monitor-name-button"
            onClick={() => setEditingName(true)}
            title="Click to rename"
          >
            {win.name}
          </button>
        )}
        <div className="monitor-label-actions">
          <span className={`tally ${win.syncing ? "synced" : "on-cycle"}`}>
            <span className="tally-light" />
            {win.syncing ? "SYNCED" : "ON CYCLE"}
          </span>
          <button
            className="monitor-delete-btn"
            onClick={handleDeleteWindow}
            disabled={busy}
            title="Remove this window"
            aria-label="Remove this window"
          >
            ✕
          </button>
        </div>
      </div>

      <div className={screenClass(current, hasPlaylist)}>
        {renderScreen(current, hasPlaylist)}
        {current && hasPlaylist && (
          <div className="monitor-scrim">
            <span className="item-title">{current.title || current.type}</span>
            <span className="timecode">
              {formatTimecode(elapsed)} / {formatTimecode(duration)}
            </span>
          </div>
        )}
      </div>

      <div className="monitor-timeline">
        <div className="monitor-timeline-fill" style={{ width: `${progressPct}%` }} />
      </div>

      {error && <div className="form-error monitor-error">{error}</div>}

      <div className="monitor-footer">
        <button className="add-media-toggle" onClick={() => setShowPlaylist((v) => !v)}>
          {showPlaylist ? "Hide" : "Show"} playlist ({win.playlist.length})
        </button>

        {showPlaylist && (
          <ul className="playlist-list">
            {win.playlist.length === 0 && (
              <li className="playlist-empty">No media yet — add some below.</li>
            )}
            {win.playlist.map((item) => (
              <li
                key={item.id}
                className={`playlist-row ${current?.id === item.id ? "is-current" : ""}`}
              >
                <span className="playlist-type">{item.type}</span>
                <span className="playlist-title">{item.title || item.id}</span>
                <span className="playlist-duration">{item.duration_seconds}s</span>
                <button
                  className="playlist-remove-btn"
                  onClick={() => handleRemoveMedia(item.id)}
                  disabled={busy}
                  title="Remove this item"
                  aria-label="Remove this item"
                >
                  ✕
                </button>
              </li>
            ))}
          </ul>
        )}

        {!addingMedia ? (
          <button className="add-media-toggle" onClick={() => setAddingMedia(true)}>
            + Add media to this window
          </button>
        ) : (
          <AddMediaForm
            windowId={win.window_id}
            onAdded={() => {
              setAddingMedia(false);
              onChanged();
            }}
            onCancel={() => setAddingMedia(false)}
          />
        )}
      </div>
    </div>
  );
}

function screenClass(current, hasPlaylist) {
  if (!hasPlaylist || !current) return "monitor-screen empty";
  if (current.type === "blank") return "monitor-screen blank";
  return "monitor-screen";
}

function renderScreen(current, hasPlaylist) {
  if (!hasPlaylist || !current) {
    return <span className="empty-label">NO PLAYLIST CONFIGURED</span>;
  }
  if (current.type === "image") {
    return <img src={current.url} alt={current.title || "media"} />;
  }
  if (current.type === "video") {
    return (
      <video
        key={current.id}
        src={current.url}
        autoPlay
        muted
        loop
        playsInline
      />
    );
  }
  return <span className="blank-label">BLANK</span>;
}