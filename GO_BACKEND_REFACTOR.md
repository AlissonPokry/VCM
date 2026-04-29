# Go Backend Refactor Change Log

## Summary

The `server/` backend was replaced from Node.js/Express/Knex/Multer to Go/chi/database/sql while preserving the frontend API contract. The frontend remains unchanged except for the allowed `partial` status constant.

## Files Added

Go backend:

- `server/main.go`
- `server/go.mod`
- `server/go.sum`
- `server/config/config.go`
- `server/db/db.go`
- `server/db/migrate.go`
- `server/db/migrations/001_create_videos.sql`
- `server/db/migrations/002_create_tags.sql`
- `server/db/migrations/003_create_activity_log.sql`
- `server/models/video.go`
- `server/models/tag.go`
- `server/models/activity.go`
- `server/models/execution.go`
- `server/handlers/videos.go`
- `server/handlers/tags.go`
- `server/handlers/analytics.go`
- `server/handlers/activity.go`
- `server/handlers/n8n.go`
- `server/services/video_service.go`
- `server/services/tag_service.go`
- `server/services/activity_service.go`
- `server/services/analytics_service.go`
- `server/services/thumbnail_service.go`
- `server/services/platform_dispatcher.go`
- `server/middleware/cors.go`
- `server/middleware/n8n_auth.go`
- `server/middleware/error_handler.go`

Documentation:

- `GO_BACKEND_REFACTOR.md`

## Files Modified

- `.env.example`: replaced `NODE_ENV` with `APP_ENV`; added `PLATFORM_DISPATCH_TIMEOUT_SECONDS`.
- `package.json`: replaced Node server scripts with Go run/build/migrate scripts; root workspace now targets the Vue client.
- `client/src/constants/index.js`: added `partial` status and orange badge classes.
- `DOCUMENTATION.md`: rewritten backend sections for Go, CGO, FFmpeg, migrations, concurrency, security, and operations.
- `README.md`: updated quick-start prerequisites and backend notes.

## Files Removed

All previous Node backend files under `server/` were removed, including Express routes, Knex migrations, Multer upload middleware, and Node services.

## API Compatibility

Routes preserved:

- `GET /api/videos`
- `GET /api/videos/:id`
- `POST /api/videos/upload`
- `PATCH /api/videos/:id`
- `PATCH /api/videos/:id/status`
- `POST /api/videos/bulk`
- `DELETE /api/videos/:id`
- `GET /api/videos/:id/thumbnail`
- `GET /api/videos/:id/file`
- `GET /api/tags`
- `GET /api/analytics/summary`
- `GET /api/analytics/heatmap`
- `GET /api/activity`
- `GET /api/n8n/queue`
- `POST /api/n8n/webhook/posted`
- `POST /api/n8n/webhook/failed`
- `GET /api/health`
- `GET /api/health/n8n`

Response envelopes preserved:

```json
{ "success": true, "data": {}, "meta": {} }
```

```json
{ "error": true, "message": "Message", "code": "STABLE_CODE" }
```

## Architecture Changes

| Old | New |
|---|---|
| Express router | chi v5 router |
| Knex + better-sqlite3 | `database/sql` + `mattn/go-sqlite3` |
| Knex JS migrations | ordered SQL migrations run by Go |
| Multer upload middleware | `net/http` multipart upload handling |
| fluent-ffmpeg + bundled ffmpeg-static | ffmpeg-go + system FFmpeg |
| Node async callbacks | Go services with context-aware SQL |
| Sequential posting behavior | parallel goroutine dispatcher for platform fan-out |

## Database Changes

No table or column changes were made. SQLite setup now applies:

- `PRAGMA journal_mode = WAL`
- `PRAGMA foreign_keys = ON`
- `PRAGMA synchronous = NORMAL`
- `PRAGMA cache_size = -64000`
- `PRAGMA busy_timeout = 5000`

The Go DB pool is capped at one open connection to avoid SQLite write contention.

## Parallel Dispatcher

`server/services/platform_dispatcher.go` implements:

- platform resolution: `all` -> `instagram`, `tiktok`, `youtube`
- one goroutine per target platform
- buffered result channel sized to platform count
- `sync.WaitGroup` close coordination
- panic recovery in every spawned goroutine
- per-platform timeout handling
- stub platform post with clear log line and TODO

When all platform results return, the n8n posted webhook appends one execution-log entry per result and sets final status to `posted` or `partial`.

## Security Changes

- n8n auth moved to Go middleware validating `x-n8n-secret`.
- CORS is configured from `CORS_ORIGIN`.
- Upload MIME type is detected server-side from file bytes.
- SQL uses parameterized queries.
- File-serving paths are constrained to upload/thumbnail directories.
- Goroutine panics are recovered and logged.

## Operational Requirements

Go backend needs:

- Go 1.22+
- GCC or equivalent C compiler for CGO
- FFmpeg on `PATH` for thumbnail extraction

If FFmpeg is missing, the server logs a warning and disables thumbnail work without rejecting uploads.

## Verification Commands

```bash
go version
gcc --version
ffmpeg -version
npm install
npm run migrate
npm run build:server
npm run dev
```

Endpoint smoke tests:

```bash
curl http://localhost:3001/api/health
curl http://localhost:3001/api/videos
curl -H "x-n8n-secret: change_me_before_production" http://localhost:3001/api/n8n/queue
```

## Known Deviations

- The platform API integration remains a stub by design. Replace `postToPlatform` in `server/services/platform_dispatcher.go` with real API clients.
- `server/go.sum` could not be generated in this environment because the local shell does not have Go installed. Running `go mod tidy` or `npm run build:server` on a machine with Go will populate it.

## Replacing `postToPlatform`

To add real posting:

1. Keep `DispatchToPlatforms` unchanged.
2. Replace only `postToPlatform`.
3. Use a per-platform API client with request timeout no longer than the provided `timeout`.
4. Return `PlatformResult{Success: false, Error: "..."}` for API failures.
5. Never panic with secrets or response payloads in the error message.
6. Keep platform credentials in environment variables or a secret manager, not in code.
