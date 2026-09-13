import { useState } from "react";
import AddMediaForm from "./AddMediaForm";
import { formatTimecode } from "../format";

export default function MonitorWindow({ win, onPlaylistChanged }) {
  const [addingMedia, setAddingMedia] = useState(false);
  const current = win.current_item;
  const hasPlaylist = win.playlist && win.playlist.length > 0;
  const elapsed = win.elapsed_in_item_seconds || 0;
  const duration = current?.duration_seconds || 0;
  const progressPct = duration > 0 ? Math.min(100, (elapsed / duration) * 100) : 0;

  return (
    <div className={`monitor ${win.syncing ? "is-synced" : ""}`}>
      <div className="monitor-label-plate">
        <span className="monitor-name">{win.name}</span>
        <span className={`tally ${win.syncing ? "synced" : "on-cycle"}`}>
          <span className="tally-light" />
          {win.syncing ? "SYNCED" : "ON CYCLE"}
        </span>
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

      <div className="monitor-footer">
        {!addingMedia ? (
          <button className="add-media-toggle" onClick={() => setAddingMedia(true)}>
            + Add media to this window
          </button>
        ) : (
          <AddMediaForm
            windowId={win.window_id}
            onAdded={(updatedWin) => {
              setAddingMedia(false);
              onPlaylistChanged(updatedWin);
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
