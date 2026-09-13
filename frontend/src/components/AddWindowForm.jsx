import { useState } from "react";
import { createWindow } from "../../api";

export default function AddWindowForm({ onCreated }) {
  const [name, setName] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState(null);

  async function handleSubmit(e) {
    e.preventDefault();
    if (!name.trim()) {
      setError("Give the window a name.");
      return;
    }
    setSubmitting(true);
    setError(null);
    try {
      const win = await createWindow(name.trim());
      setName("");
      onCreated(win);
    } catch (err) {
      setError(err.message);
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="add-window-tile">
      <form onSubmit={handleSubmit}>
        <p>Add another display window</p>
        <input
          className="field"
          placeholder="Window name, e.g. Window 3"
          value={name}
          onChange={(e) => setName(e.target.value)}
        />
        {error && <div className="form-error">{error}</div>}
        <button type="submit" className="btn btn-primary" disabled={submitting}>
          {submitting ? "Creating…" : "Create window"}
        </button>
      </form>
    </div>
  );
}
