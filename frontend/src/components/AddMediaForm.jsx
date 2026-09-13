import { useState } from "react";
import { addMedia } from "../../api";

const TYPE_DEFAULTS = {
  image: 5,
  video: 10,
  blank: 3,
};

export default function AddMediaForm({ windowId, onAdded, onCancel }) {
  const [type, setType] = useState("image");
  const [url, setUrl] = useState("");
  const [title, setTitle] = useState("");
  const [duration, setDuration] = useState(TYPE_DEFAULTS.image);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  function handleTypeChange(nextType) {
    setType(nextType);
    setDuration(TYPE_DEFAULTS[nextType]);
  }

  async function handleSubmit(e) {
    e.preventDefault();
    if (type !== "blank" && !url.trim()) {
      setError("A URL is required for image and video items.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const win = await addMedia(windowId, {
        type,
        url: type === "blank" ? "" : url.trim(),
        title: title.trim(),
        duration_seconds: Number(duration) || TYPE_DEFAULTS[type],
      });
      onAdded(win);
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <form className="add-media-form" onSubmit={handleSubmit}>
      <div className="row">
        <select
          className="field"
          value={type}
          onChange={(e) => handleTypeChange(e.target.value)}
        >
          <option value="image">Image</option>
          <option value="video">Video</option>
          <option value="blank">Blank</option>
        </select>
        <input
          className="field"
          style={{ width: 64, flex: "none" }}
          type="number"
          min="1"
          value={duration}
          onChange={(e) => setDuration(e.target.value)}
          aria-label="Duration in seconds"
        />
      </div>
      {type !== "blank" && (
        <input
          className="field"
          placeholder={type === "image" ? "Image URL" : "Video URL"}
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
      )}
      <input
        className="field"
        placeholder="Title (optional)"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
      />
      {error && <div className="form-error">{error}</div>}
      <div className="row">
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? "Adding…" : "Add to playlist"}
        </button>
        <button type="button" className="btn" onClick={onCancel} disabled={submitting}>
          Cancel
        </button>
      </div>
    </form>
  );
}
