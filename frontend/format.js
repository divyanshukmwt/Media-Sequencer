export function formatTimecode(totalSeconds) {
  const s = Math.max(0, Math.floor(totalSeconds || 0));
  const mm = String(Math.floor(s / 60)).padStart(2, "0");
  const ss = String(s % 60).padStart(2, "0");
  return `${mm}:${ss}`;
}
