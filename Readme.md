# Media Sequencer — Multi-Window Media Sequencer with Sync Playback

A full-stack app where multiple independent display "windows" each continuously
loop their own configured media playlist (images, videos, or blank slots), and
a **sync/take** control can push one selected item to every window at the same
instant — then every window resumes its own sequence exactly where it left off.

Built for the Backend Intern take-home assignment: **React frontend + Golang
backend**.

---

## Table of contents

1. [Assignment requirements](#assignment-requirements)
2. [Features](#features)
3. [Tech stack](#tech-stack)
4. [Project structure](#project-structure)
5. [Prerequisites](#prerequisites)
6. [Environment variables](#environment-variables)
7. [Database setup](#database-setup)
8. [Backend setup](#backend-setup)
9. [Frontend setup](#frontend-setup)
10. [Seed data](#seed-data)
11. [How to run the complete project](#how-to-run-the-complete-project)
12. [How to test the application](#how-to-test-the-application)
13. [API documentation](#api-documentation)
14. [Playback architecture](#playback-architecture)
15. [Sync behavior](#sync-behavior)
16. [Deployment](#deployment)
17. [Assumptions](#assumptions)
18. [Architecture tradeoffs](#architecture-tradeoffs)
19. [Troubleshooting](#troubleshooting)

---

## Assignment requirements

- Each window has its own media list; the total play size per window is
  treated as **5 hours**, within which its list repeats continuously.
- Blank is only shown where it's an actual configured playlist item — the
  rest of the cycle never silently goes blank.
- A **sync** action can push one selected media item to every window at the
  same time; after the configured sync duration, every window resumes its own
  sequence without losing its playlist configuration.
- React frontend, Golang backend, persistent storage, dynamic playlist
  updates, both apps deployed.

Full original spec: see the assignment PDF supplied alongside this project.

## Features

- Multiple display windows ("monitors"), each with its own independently
  scheduled, continuously repeating playlist.
- Image, video, and blank media types.
- Dynamic playlist updates — add media to any window at runtime, no restart
  needed.
- A global **sync/take** control: pick any media item currently in any
  window's playlist, push it to all windows for a configurable duration, then
  every window automatically resumes its own schedule.
- Persistent storage in MongoDB — playlists and sync state survive restarts.
- Seed data matching the assignment's example (`Window 1: M1→M2→M3→repeat`,
  `Window 2: M4→M5→repeat`), loaded automatically on first run.
- All playback/sync timing is computed **server-side** — the frontend just
  renders whatever the backend says is current, polled once a second.

## Tech stack

| Layer     | Choice                                                             |
|-----------|---------------------------------------------------------------------|
| Frontend  | React 19 + Vite, plain CSS (no UI framework)                        |
| Backend   | Go 1.22, standard library `net/http` (native method+path routing)   |
| Database  | MongoDB (official `go.mongodb.org/mongo-driver`)                    |
| Testing   | Go's built-in `testing` package (24 tests, no external framework)   |

No web framework, ORM, or state-management library was added — the surface
area of this app doesn't need one, and the assignment rewards "clean and
simple" over "impressive-looking dependency list."

## Project structure

```
media-sequencer/
├── backend/
│   ├── main.go            # entrypoint: config, Mongo connect, seed, serve
│   ├── config.go          # env var loading
│   ├── models.go          # Window, MediaItem, SyncState structs + constants
│   ├── scheduler.go       # pure playback/cycle/sync math (the core logic)
│   ├── scheduler_test.go  # 13 unit tests for the scheduling math
│   ├── store.go           # Store interface + MongoDB implementation
│   ├── handlers.go        # HTTP handlers, routing, CORS
│   ├── handlers_test.go   # 11 HTTP-layer tests (via an in-memory fake store)
│   ├── fake_store_test.go # in-memory Store used only by tests
│   ├── seed.go            # idempotent seed data loader
│   ├── Dockerfile
│   ├── .env.example
│   ├── go.mod / go.sum
├── frontend/
│   ├── src/
│   │   ├── main.jsx
│   │   ├── App.jsx            # polls /api/state, renders the monitor wall
│   │   ├── api.js              # fetch wrapper, configurable API base URL
│   │   ├── format.js            # timecode formatting
│   │   ├── index.css            # design tokens + all styling
│   │   └── components/
│   │       ├── MonitorWindow.jsx  # one window: screen, tally, timeline, add-media
│   │       ├── AddMediaForm.jsx   # add a media item to a window's playlist
│   │       ├── AddWindowForm.jsx  # create a new window
│   │       └── SyncPanel.jsx      # media picker + "TAKE TO ALL MONITORS"
│   ├── index.html
│   ├── .env.example
│   └── package.json
├── README.md               (this file)
└── ASSIGNMENT_GUIDE.md      (implementation deep-dive, commit plan, interview prep)
```

## Prerequisites

- Go 1.22+
- Node.js 18+ and npm
- A MongoDB instance — either:
  - a local `mongod` (easiest for development), or
  - a free MongoDB Atlas cluster (needed for the deployed version anyway)

## Environment variables

**`backend/.env.example`** (copy to `backend/.env` or set these as real
environment variables — the Go app does not auto-load `.env` files, see
[Backend setup](#backend-setup)):

| Variable          | Default                     | Meaning                                    |
|-------------------|------------------------------|---------------------------------------------|
| `PORT`            | `8080`                       | port the Go server listens on               |
| `MONGODB_URI`      | `mongodb://localhost:27017`| MongoDB connection string                   |
| `MONGODB_DB`       | `mediasequencer`           | database name (auto-created)                |
| `FRONTEND_ORIGIN`  | `*`                          | value for `Access-Control-Allow-Origin`     |

**`frontend/.env.example`**:

| Variable              | Default                    | Meaning                        |
|-----------------------|------------------------------|---------------------------------|
| `VITE_API_BASE_URL`   | `http://localhost:8080`   | base URL of the backend API     |

## Database setup

**Option A — local MongoDB (fastest for development):**

```bash
# macOS (Homebrew)
brew tap mongodb/brew && brew install mongodb-community
brew services start mongodb-community

# Ubuntu/Debian: follow MongoDB's official apt-repo instructions for your
# release, then:
sudo systemctl start mongod
```

Leave `MONGODB_URI=mongodb://localhost:27017` (the default).

**Option B — MongoDB Atlas free tier (needed for deployment regardless):**

1. Create a free account at mongodb.com/atlas.
2. Create a free **M0** cluster.
3. Under *Database Access*, create a user with a password.
4. Under *Network Access*, add `0.0.0.0/0` (or your deploy host's IP) so
   Render/your backend can reach it.
5. Click *Connect → Drivers*, copy the `mongodb+srv://...` connection string,
   and use it as `MONGODB_URI` (fill in the real password).

The database and collections (`windows`, `sync_state`) are created
automatically on first write — nothing to run manually.

## Backend setup

```bash
cd backend
cp .env.example .env   # then edit MONGODB_URI if not using local default

go mod download        # see note below if this fails
go run .
```

You should see:

```
connected to MongoDB (db=mediasequencer)
seed: inserted window-1 (M1,M2,M3) and window-2 (M4,M5)
media-sequencer backend listening on :8080 (frontend origin: *)
```

> **Note on `go mod download` / module resolution.** This project's
> `go.mod` uses `replace` directives that point `go.mongodb.org/mongo-driver`
> and a few `golang.org/x/...` packages at their canonical GitHub source
> (e.g. `go.mongodb.org/mongo-driver` **is** `github.com/mongodb/mongo-go-driver`
> — that's just its vanity import path). This was necessary because the
> sandbox this project was built in could reach `github.com` but not
> `proxy.golang.org`. On a normal machine with unrestricted internet this
> isn't required, but it's harmless to leave in — it just makes module
> resolution work identically on any network. If you ever see a `go get`
> error mentioning a `golang.org/x/...` package, it means Go needs one more
> transitive dependency resolved; add a matching
> `replace golang.org/x/<pkg> => github.com/golang/<pkg> <version>` line to
> `go.mod` (see [ASSIGNMENT_GUIDE.md](./ASSIGNMENT_GUIDE.md) for the full
> explanation of how these were derived).

Since the Go program reads real environment variables (not `.env` files
directly), either `export` the values from `.env` yourself, or use a tool
like [`direnv`](https://direnv.net/) or run:

```bash
export $(grep -v '^#' .env | xargs) && go run .
```

## Frontend setup

```bash
cd frontend
cp .env.example .env   # edit VITE_API_BASE_URL if backend isn't on localhost:8080

npm install
npm run dev
```

Open the printed local URL (typically `http://localhost:5173`).

## Seed data

On first run against an empty database, the backend automatically inserts:

- **Window 1**: M1 (image, 8s) → M2 (video, 15s) → M3 (blank, 5s) → repeat
- **Window 2**: M4 (image, 10s) → M5 (video, 12s) → repeat

This matches the assignment's example pattern. Seeding is idempotent — it
only runs when the `windows` collection is empty, so restarting the backend
never duplicates data.

## How to run the complete project

1. Clone the repository.
2. Start MongoDB (local `mongod`, or have your Atlas URI ready).
3. `cd backend && cp .env.example .env` (edit `MONGODB_URI` if needed).
4. `cd backend && go run .` — leave running.
5. In a second terminal: `cd frontend && cp .env.example .env`.
6. `cd frontend && npm install && npm run dev`.
7. Open the app in your browser.
8. You should see two monitor tiles (Window 1, Window 2) already cycling
   through the seeded playlist.
9. Click **"+ Add media to this window"** on either monitor to add a new
   image/video/blank item — it appears in that window's schedule immediately.
10. In the **SYNC / TAKE** panel at the bottom, click any media chip, set a
    duration, and click **TAKE TO ALL MONITORS** — every window switches to
    that item immediately, then resumes its own sequence when the duration
    ends.

## How to test the application

**Automated tests (backend):**

```bash
cd backend
go test ./... -v
```

24 tests covering:
- the cycle/loop/wraparound math (13 tests in `scheduler_test.go`)
- the full HTTP API contract — add media, trigger sync, validation, 404s (11
  tests in `handlers_test.go`, run against an in-memory fake store so no
  live database is needed to run them)

**Manual smoke test (full stack):**

1. Start the backend and frontend as above.
2. Confirm both seeded windows are visibly cycling through their items.
3. Add a media item to Window 1; confirm it appears in that window's rotation
   within one polling cycle (≤1s) and that Window 2 is unaffected.
4. Refresh the browser tab — confirm the same playlists and roughly the same
   playback position reload (proves persistence, since position is derived
   from `cycle_start_at` stored in MongoDB, not from browser memory).
5. Trigger a sync on some media item with a short duration (e.g. 5s); confirm
   **every** window switches to it immediately, and that both windows'
   original playlists are still intact underneath (visible again once the
   sync duration elapses).
6. Stop and restart the backend process; confirm playlists and playback
   position are unaffected (data is in MongoDB, not in server memory).

> This project was built and unit-tested in a sandboxed environment without
> a running MongoDB server available, so the automated tests above run
> against an in-memory fake rather than live MongoDB (see
> [ASSIGNMENT_GUIDE.md](./ASSIGNMENT_GUIDE.md) for why, and what to double
> check). Steps 1–6 above are what to run yourself against a real MongoDB to
> confirm the full stack end-to-end before submitting.

## API documentation

Base path: `/api`. All responses are JSON. Timestamps are RFC3339 UTC.

| Method | Path                          | Purpose                                            |
|--------|-------------------------------|-----------------------------------------------------|
| GET    | `/health`                     | liveness check                                      |
| GET    | `/api/state`                  | **primary endpoint** — all windows with their currently-resolved item, plus sync status. Polled every 1s by the frontend. |
| GET    | `/api/windows`                | list all windows (raw playlists, no resolved item)  |
| POST   | `/api/windows`                | create a window — body: `{ "name": "Window 3" }`   |
| GET    | `/api/windows/{id}`           | get one window                                       |
| POST   | `/api/windows/{id}/media`     | add a media item — body: `{ "type": "image"\|"video"\|"blank", "url": "...", "title": "...", "duration_seconds": 8 }` (`url` required for image/video; `duration_seconds` defaults per type if omitted) |
| GET    | `/api/sync`                   | current sync status                                  |
| POST   | `/api/sync`                   | trigger a sync — body: `{ "media_id": "m2", "duration_seconds": 10 }` (`duration_seconds` defaults to 10) |

Example `GET /api/state` response shape:

```json
{
  "server_time": "2026-09-12T10:00:00Z",
  "windows": [
    {
      "window_id": "window-1",
      "name": "Window 1",
      "playlist": [ { "id": "m1", "type": "image", "url": "...", "title": "M1 - ...", "duration_seconds": 8 } ],
      "cycle_start_at": "2026-09-12T05:00:00Z",
      "current_item": { "id": "m2", "type": "video", "url": "...", "duration_seconds": 15 },
      "elapsed_in_item_seconds": 3,
      "syncing": false
    }
  ],
  "sync": { "active": false, "remaining_seconds": 0 }
}
```

## Playback architecture

Every window's playback position is **computed on demand**, not tracked as
mutable server state. Each window stores a `cycle_start_at` timestamp
(set once, at creation) and its playlist (each item has a
`duration_seconds`). To find what's showing right now:

1. `elapsed = now - cycle_start_at`
2. `cyclePos = elapsed mod 5 hours` — this is the literal "5-hour cycle"
   from the spec; it resets to 0 every 5 hours.
3. `playlistPos = cyclePos mod (sum of the playlist's durations)` — within
   the current 5-hour cycle, the (much shorter) playlist repeats
   back-to-back.
4. Walk the playlist's cumulative durations to find which item covers
   `playlistPos`.

This is a pure function of `(playlist, cycle_start_at, now)`. Nothing is
mutated, no background goroutine ticks a "current index," and a server
restart never loses playback position — any number of clients computing this
independently get the identical answer. See `CurrentPlaylistItem` in
`backend/scheduler.go`.

## Sync behavior

Sync is a single global override record: `{ active, media, started_at,
duration_seconds }`, stored in its own MongoDB document (not per-window).

- `POST /api/sync` looks up the requested `media_id` across every window's
  playlist, then writes that override record.
- Every `GET /api/state` call resolves, per window: *if the sync is active
  and `now` is before `started_at + duration_seconds`, show the synced
  item; otherwise show the window's own scheduled item.* This check
  (`ResolveDisplay` in `scheduler.go`) is what makes "every window shows the
  same item at the same time" true — they're all reading the same one sync
  record.
- Expiry needs no timer or cleanup job — "is the sync still active" is just
  a timestamp comparison done at read time. Once `now` passes the expiry, the
  very next `/api/state` poll naturally reports each window's own item
  again, computed via the same cycle math as if sync had never happened —
  because a window's `cycle_start_at` is never touched by sync at all.

## Deployment

**This project was not deployed from the environment it was built in** — that
sandbox has network access to package registries only (npm, Go modules via
GitHub, etc.), not to any hosting provider. The steps below are exact and
tested-as-written; you'll need to run them yourself.

**1. Database — MongoDB Atlas** (free M0 tier): see
[Database setup](#database-setup) Option B above. Keep the connection string.

**2. Backend — Render** (free web service, supports Docker):

1. Push this repo to GitHub (see the commit plan in
   [ASSIGNMENT_GUIDE.md](./ASSIGNMENT_GUIDE.md)).
2. On [render.com](https://render.com), **New → Web Service**, connect the
   repo, set **Root Directory** to `backend`.
3. Environment: **Docker** (Render will use `backend/Dockerfile`).
4. Add environment variables: `MONGODB_URI` (your Atlas string),
   `MONGODB_DB=mediasequencer`, `FRONTEND_ORIGIN` (set this **after** step 3
   below, once you know the frontend's URL — `*` works temporarily).
5. Deploy. Note the resulting URL, e.g. `https://media-sequencer-backend.onrender.com`.
6. Confirm `https://<that-url>/health` returns `{"status":"ok"}`.

   > Free Render web services spin down after 15 minutes of inactivity and
   > take ~30–60s to wake on the next request — the first request after
   > idling will be slow. This is a free-tier limitation, not a bug.

**3. Frontend — Vercel** (free static hosting):

1. On [vercel.com](https://vercel.com), **Add New → Project**, import the
   repo, set **Root Directory** to `frontend`.
2. Framework preset: Vite (auto-detected).
3. Add environment variable `VITE_API_BASE_URL` = your Render backend URL
   from step 2.
4. Deploy. Note the resulting URL, e.g. `https://media-sequencer.vercel.app`.

**4. Close the loop:** go back to the Render backend's environment variables
and set `FRONTEND_ORIGIN` to your exact Vercel URL (e.g.
`https://media-sequencer.vercel.app`, no trailing slash), then redeploy the
backend so CORS only allows your real frontend.

**Live URLs:** _fill in after you complete the steps above —_
- Frontend: `<your Vercel URL>`
- Backend: `<your Render URL>`

## Assumptions

- "Windows" are panels within a single React page (a monitor wall), not
  separate OS-level browser windows — this matches how the assignment's own
  diagram reads (`Window 1: M1→M2→M3→repeat` alongside `Window 2`, displayed
  together) and is the standard interpretation for this kind of
  digital-signage assignment.
- Every media item — image, video, *and* blank — carries an explicit
  `duration_seconds` used for scheduling, rather than videos being timed by
  their own natural length. This keeps every window's schedule a pure,
  deterministic function of time (see [Playback architecture](#playback-architecture)),
  so any number of clients/tabs stay in agreement without a shared video
  clock or websocket.
- Sync is global (one at a time, across all windows), matching "every window
  should display M2 at the same time" — not per-window.
- The 5-hour cycle wraps by resetting `playlistPos` to 0 at the exact 5-hour
  mark, even if that cuts an item short (see the wraparound test case in
  `scheduler_test.go`). The assignment doesn't specify remainder handling;
  this is the simplest literal reading of "total play size... treated as 5
  hours."
- A media item, once added to a window's playlist, gets a permanent unique
  ID; sync targets that exact ID (found by scanning all windows), not a
  freeform URL — this keeps `/api/sync` a single simple call.

## Architecture tradeoffs

- **No websockets.** The frontend polls `/api/state` every second instead of
  the backend pushing updates. Simpler, works through any host/proxy without
  special config, and a 1-second worst-case sync latency is imperceptible
  for this use case. Tradeoff: slightly more HTTP traffic than a push model.
- **Schedule computed at read time, not stored as a "current index."** Makes
  the system stateless and trivially consistent across clients and restarts,
  at the cost of a few extra CPU cycles per request (negligible at this
  scale).
- **Global (not per-window) sync record.** Matches the spec directly and is
  much simpler than per-window sync state, at the cost of not supporting
  "sync only windows A and B" (not required by the assignment).
- **No framework/ORM on either side.** Faster to review, fewer moving parts,
  fewer places for a hidden bug to hide — appropriate for an app this size.
  Tradeoff: a much larger app would eventually want one.
- **Free-tier hosting caveats:** Render's free web services have an
  *ephemeral* filesystem and spin down when idle — irrelevant here since all
  state lives in MongoDB Atlas, not on local disk, but worth knowing if you
  ever add local file storage.

## Troubleshooting

| Symptom | Likely cause / fix |
|---|---|
| Backend exits immediately with a Mongo connection error | `MONGODB_URI` wrong, `mongod` not running, or (for Atlas) your IP isn't allow-listed under Network Access |
| Frontend shows "backend unreachable" | Backend isn't running, wrong `VITE_API_BASE_URL`, or CORS: `FRONTEND_ORIGIN` on the backend doesn't match the frontend's actual origin |
| `go mod download` fails on a `golang.org/x/...` package | See the module resolution note under [Backend setup](#backend-setup) — add one more `replace` line |
| Video doesn't play in a monitor tile | Browser autoplay policies require video to be muted for autoplay — this app already sets `muted`; check the browser console for a blocked-media error, or that the video URL is directly playable (not an HTML page) |
| Playlist added but monitor doesn't update | Wait up to 1s (polling interval); if it never updates, check the browser console/network tab for a failed `POST .../media` request |
| Sync doesn't affect a window | Confirm `/api/state`'s `sync.active` is `true` and hasn't already expired (`remaining_seconds`) |
