export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080";

async function request(path, options = {}) {
  const res = await fetch(`${API_BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    let message = `Request failed (${res.status})`;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
    }
    throw new Error(message);
  }
  if (res.status === 204) return null;
  return res.json();
}

export function getState() {
  return request("/api/state");
}

export function getWindows() {
  return request("/api/windows");
}

export function createWindow(name) {
  return request("/api/windows", {
    method: "POST",
    body: JSON.stringify({ name }),
  });
}

export function addMedia(windowId, media) {
  return request(`/api/windows/${encodeURIComponent(windowId)}/media`, {
    method: "POST",
    body: JSON.stringify(media),
  });
}

export function removeMedia(windowId, mediaId) {
  return request(
    `/api/windows/${encodeURIComponent(windowId)}/media/${encodeURIComponent(mediaId)}`,
    { method: "DELETE" }
  );
}

export function renameWindow(windowId, name) {
  return request(`/api/windows/${encodeURIComponent(windowId)}`, {
    method: "PATCH",
    body: JSON.stringify({ name }),
  });
}

export function deleteWindow(windowId) {
  return request(`/api/windows/${encodeURIComponent(windowId)}`, {
    method: "DELETE",
  });
}

export function triggerSync(mediaId, durationSeconds) {
  return request("/api/sync", {
    method: "POST",
    body: JSON.stringify({ media_id: mediaId, duration_seconds: durationSeconds }),
  });
}