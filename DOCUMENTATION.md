# V.C.M (Vertical Content Manager) Documentation

## 1. Project Overview

V.C.M is a local-first vertical video scheduling dashboard for Reels, TikTok, and YouTube Shorts. It uploads video files, schedules posts, tracks publishing status, records activity, exposes analytics, and provides secret-protected n8n automation endpoints.

The backend is now a Go service. The frontend API contract remains unchanged: `client/src/api/*.js` still calls the same routes and receives the same response envelopes.

## 2. Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Vue 3, Pinia, Vue Router, Tailwind CSS, Axios |
| Backend | Go 1.22+, `net/http`, `chi` v5 |
| Database | SQLite via `database/sql` and `mattn/go-sqlite3` |
| Migrations | `golang-migrate/migrate` with ordered SQL files |
| Video Processing | `u2takey/ffmpeg-go` over system FFmpeg |
| Config | `joho/godotenv` plus typed Go config |
| IDs | `google/uuid` for uploaded filenames |

SQLite remains the storage engine because this project is local-first and single-user. The Go backend enables stronger concurrency, simpler deployment binaries, and parallel platform dispatch for multi-platform posting.

## 3. Prerequisites

- Node.js v20+ and npm v9+ for the Vue client and root scripts.
- Go 1.22+ for the backend.
- GCC or another CGO-compatible C compiler. `mattn/go-sqlite3` requires CGO.
- FFmpeg installed on `PATH` for thumbnail and duration extraction. If FFmpeg is missing, uploads still work, but thumbnail processing is disabled with a startup warning.
- Writable project directory for SQLite, uploads, thumbnails, and Go build output.

## 4. Installation & Setup

```bash
npm install
cp .env.example .env
npm run migrate
npm run dev
```

Root scripts:

| Script | Command |
|---|---|
| `npm run dev` | Starts Go API and Vite client together |
| `npm run dev:server` | Runs `go run main.go` inside `server/` |
| `npm run dev:client` | Runs Vite inside `client/` |
| `npm run build` | Builds Go server binary and Vue client |
| `npm run migrate` | Runs Go migrations only, then exits |

## 5. Environment Variables

| Variable | Default | Description |
|---|---:|---|
| `PORT` | `3001` | Go API port |
| `APP_ENV` | `development` | Runtime mode; production skips `.env` loading |
| `DB_PATH` | `./server/db/reel_queue.sqlite` | SQLite file path |
| `UPLOAD_DIR` | `./server/uploads` | Uploaded video storage |
| `THUMBNAIL_DIR` | `./server/thumbnails` | Extracted JPEG thumbnail storage |
| `MAX_FILE_SIZE_MB` | `500` | Upload limit |
| `N8N_WEBHOOK_SECRET` | `change_me_before_production` | Required `x-n8n-secret` value |
| `CORS_ORIGIN` | `http://localhost:5173` | Allowed frontend origin |
| `PLATFORM_DISPATCH_TIMEOUT_SECONDS` | `60` | Per-platform dispatch timeout |

## 6. Architecture

```text
Vue SPA + Pinia + Axios
          |
          v
Go chi REST API ---- n8n polling/webhooks
          |
          v
database/sql + SQLite + local uploads/thumbnails
```

Backend structure:

```text
server/
  main.go
  config/
  db/
  handlers/
  middleware/
  models/
  services/
  uploads/
  thumbnails/
```

Handlers own HTTP parsing and response envelopes. Services own business rules, DB writes, activity logging, thumbnail processing, analytics, and n8n platform dispatch. DB access uses parameterized SQL through `database/sql`.

## 7. Database

Migrations are SQL files under `server/db/migrations`:

- `001_create_videos.sql`
- `002_create_tags.sql`
- `003_create_activity_log.sql`

The schema is unchanged from the original application: `videos`, `tags`, `video_tags`, and `activity_log`. The Go DB opener enables WAL mode, foreign keys, normal synchronous mode, a 64 MB page cache, a 5 second busy timeout, and a single open SQLite connection.

## 8. API Contract

Success responses:

```json
{ "success": true, "data": {}, "meta": {} }
```

Error responses:

```json
{ "error": true, "message": "Message", "code": "STABLE_CODE" }
```

Videos:

- `GET /api/videos`
- `GET /api/videos/:id`
- `POST /api/videos/upload`
- `PATCH /api/videos/:id`
- `PATCH /api/videos/:id/status`
- `POST /api/videos/bulk`
- `DELETE /api/videos/:id`
- `GET /api/videos/:id/thumbnail`
- `GET /api/videos/:id/file`

Tags:

- `GET /api/tags`

Analytics:

- `GET /api/analytics/summary`
- `GET /api/analytics/heatmap`

Activity:

- `GET /api/activity`

n8n:

- `GET /api/n8n/queue`
- `POST /api/n8n/webhook/posted`
- `POST /api/n8n/webhook/failed`

Health:

- `GET /api/health`
- `GET /api/health/n8n`

## 9. Status Values

Valid statuses:

- `scheduled`
- `posted`
- `draft`
- `partial`

`partial` is used when a multi-platform post succeeds on at least one platform and fails on another. The frontend status constant now includes an orange badge style for this state.

## 10. Uploads And Media

Upload validation uses server-side MIME detection from the first 512 bytes and requires a `video/*` type. Uploaded files receive UUID-based filenames. `GET /api/videos/:id/file` uses `http.ServeContent`, preserving range requests so browser video seeking works.

Thumbnail processing is asynchronous. The upload response is not blocked by FFmpeg. Processing extracts one padded `640x360` JPEG frame and stores rounded duration in seconds. If processing fails, the video remains usable with `thumbnail = null` and `duration = null`.

## 11. n8n Integration

All `/api/n8n/*` routes require `x-n8n-secret`. The queue endpoint returns due scheduled videos where `status = scheduled` and `scheduled_at <= now`.

The posted webhook appends execution log entries. When the stored video platform is `all`, the Go dispatcher expands to `instagram`, `tiktok`, and `youtube`, runs each platform post concurrently with goroutines, collects results through a buffered channel, writes one unified execution log update, and sets:

- `posted` when all platforms succeed.
- `partial` when any platform fails.

Current platform posting is a stub in `server/services/platform_dispatcher.go`. It logs `STUB: would post...`, waits two seconds, and returns success unless the configured timeout is shorter.

## 12. Security Model

- Strict CORS from `CORS_ORIGIN`.
- Secret-protected n8n endpoints.
- Parameterized SQL only.
- Upload size limit and MIME allowlist.
- Path traversal prevention for served local files.
- Panic recovery in chi middleware and all backend goroutine entry points.
- API-safe error envelopes.

## 13. Observability

The API uses chi request logging. Activity events are written for uploads, edits, status changes, deletes, n8n queue pickup, n8n success, and n8n failure. Activity writes are fire-and-forget and cannot fail the primary request.

## 14. Deployment Notes

Build the backend with:

```bash
cd server
go build -o bin/reelqueue-server .
```

Production hosts must install Go build dependencies, GCC for CGO, and FFmpeg. Keep SQLite, uploads, and thumbnails on persistent storage and back them up together.

## 15. Change Reference

Detailed migration notes live in [GO_BACKEND_REFACTOR.md](GO_BACKEND_REFACTOR.md).
