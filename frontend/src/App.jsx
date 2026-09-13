import { useCallback, useEffect, useRef, useState } from "react";
import { getState } from "../api";
import MonitorWindow from "./components/MonitorWindow";
import AddWindowForm from "./components/AddWindowForm";
import SyncPanel from "./components/SyncPanel";

const POLL_INTERVAL_MS = 1000;

export default function App() {
  const [state, setState] = useState(null);
  const [connected, setConnected] = useState(true);
  const [loadError, setLoadError] = useState(null);
  const pollingRef = useRef(null);

  const fetchState = useCallback(async () => {
    try {
      const data = await getState();
      setState(data);
      setConnected(true);
      setLoadError(null);
    } catch (err) {
      setConnected(false);
      setLoadError(err.message);
    }
  }, []);

  useEffect(() => {
    fetchState();
    pollingRef.current = setInterval(fetchState, POLL_INTERVAL_MS);
    return () => clearInterval(pollingRef.current);
  }, [fetchState]);

  return (
    <div className="app">
      <header className="app-header">
        <h1 className="app-title">
          Media Sequencer <span>— control room</span>
        </h1>
        <div className="connection-status">
          <span className={`connection-dot ${connected ? "" : "offline"}`} />
          {connected ? "backend connected" : "backend unreachable"}
        </div>
      </header>

      {!state && !loadError && <div className="loading-state">Connecting to backend…</div>}

      {loadError && !state && (
        <div className="error-banner">
          Couldn't reach the backend at start-up: {loadError}. Confirm the API is running and
          VITE_API_BASE_URL is set correctly.
        </div>
      )}

      {state && (
        <>
          <div className="monitor-grid">
            {state.windows.map((win) => (
              <MonitorWindow key={win.window_id} win={win} onPlaylistChanged={fetchState} />
            ))}
            <AddWindowForm onCreated={fetchState} />
          </div>

          <SyncPanel windows={state.windows} sync={state.sync} onSynced={fetchState} />
        </>
      )}
    </div>
  );
}
