# V.C.M (Vertical Content Manager) Documentation

## 1. Project Overview

V.C.M is a single-user vertical content manager for short-form video workflows. It helps creators and operators upload video files, schedule posts, track published items, inspect activity, and connect an n8n automation that polls due videos and reports posting results.

Capabilities:
- Upload video files with metadata, schedule, platform, and tags.
- Extract thumbnails and duration with bundled FFmpeg.
- Browse scheduled and posted queues with search, platform, tag, date, and sort filters.
- Edit video metadata and status from a responsive detail panel.
- Run bulk delete, draft, and reschedule actions.
- Review analytics, heatmaps, platform counts, tag usage, and activity history.
- Expose secret-protected n8n queue and webhook endpoints.

## 2. Tech Stack

| Layer | Technology | Version Target |
|---|---|---|
| Frontend | Vue 3, Composition API, `<script setup>` | `^3.5` |
| Styling | Tailwind CSS | `^3.4` |
| Backend | Node.js + Express | Node 20, Express 4 |
| Database | SQLite via `better-sqlite3` | `^11` runtime |
| Query Builder | Knex | `^3` |
| Uploads | Multer | `1.4.5-lts` |
| Video Processing | fluent-ffmpeg + ffmpeg-static | bundled binary |
| HTTP Client | Axios | `^1` |
| State | Pinia | `^2` |
| Router | Vue Router | `^4` |
| Icons | Lucide Vue Next | `^0.468` |

SQLite was chosen because v1 is a local-first, single-user tool where simple deployment and persistent files matter more than distributed writes.

## 3. Prerequisites

- Node.js v20+
- npm v9+
- No system FFmpeg install is required; `ffmpeg-static` provides the binary.
- A writable project directory for SQLite, uploads, thumbnails, and build output.

## 4. Installation & Setup

1. Clone or open the project folder.
2. Run `npm install` from the root to install root, client, and server workspace dependencies.
3. Copy `.env.example` to `.env` and change `N8N_WEBHOOK_SECRET`.
4. Run `npm run migrate` to create SQLite tables.
5. Run `npm run dev` to start Express on `http://localhost:3001` and Vite on `http://localhost:5173`.

## 5. Environment Variables Reference

| Variable | Default | Required | Description |
|---|---:|---|---|
| `PORT` | `3001` | No | Express API port |
| `NODE_ENV` | `development` | No | Runtime mode |
| `DB_PATH` | `./server/db/reel_queue.sqlite` | No | SQLite file path |
| `UPLOAD_DIR` | `./server/uploads` | No | Uploaded video storage |
| `THUMBNAIL_DIR` | `./server/thumbnails` | No | Extracted JPEG thumbnail storage |
| `MAX_FILE_SIZE_MB` | `500` | No | Multer upload size limit |
| `N8N_WEBHOOK_SECRET` | `change_me_before_production` | Yes for n8n | Secret checked against `x-n8n-secret` |
| `CORS_ORIGIN` | `http://localhost:5173` | No | Allowed frontend origin |

## 6. Application Architecture

```text
Vue SPA + Pinia + Axios
          |
          v
Express REST API ---- n8n polling/webhooks
          |
          v
Knex + SQLite + local uploads/thumbnails
```

The frontend owns user interaction and state. The Express API owns validation, file handling, DB updates, activity logging, analytics, and n8n integration. SQLite stores metadata, tags, junction rows, and activity events. Video files and thumbnails stay on disk.

## 7. Database Schema Reference

`videos`: primary record for each upload. Stores title, description, disk filename, original filename, size, duration, thumbnail path, platform, status, schedule/post timestamps, optional n8n workflow id, JSON execution log, and timestamps.

`tags`: normalized tag names with case-insensitive uniqueness.

`video_tags`: junction table connecting videos and tags. Composite primary key prevents duplicate tag links.

`activity_log`: append-only audit events with optional `video_id`, action, detail, source, and created timestamp.

Add schema changes as new Knex migration files under `server/db/migrations`. Do not edit applied migrations.

## 8. API Reference

All success responses use `{ "success": true, "data": value, "meta": object }`. Errors use `{ "error": true, "message": string, "code": string }`.

Videos:
- `GET /api/videos`: query `status`, `platform`, `tag`, `search`, `sort`, `order`, `dateFrom`, `dateTo`; returns videos with `tags`.
- `GET /api/videos/:id`: returns one video with `tags`.
- `POST /api/videos/upload`: multipart `file`, `title`, `description`, `platform`, `scheduled_at`, `tags`; creates video, thumbnail, duration, activity.
- `PATCH /api/videos/:id`: updates metadata and tags.
- `PATCH /api/videos/:id/status`: body `{ "status": "scheduled" | "posted" | "draft" }`.
- `POST /api/videos/bulk`: body `{ "ids": [], "action": "delete" | "draft" | "reschedule", "scheduled_at": "..." }`.
- `DELETE /api/videos/:id`: deletes DB row, file, and thumbnail.
- `GET /api/videos/:id/thumbnail`: serves JPEG thumbnail.
- `GET /api/videos/:id/file`: serves the video file.

Tags:
- `GET /api/tags`: returns `{ id, name, count }` sorted by usage.

Analytics:
- `GET /api/analytics/summary`: returns all KPI totals and platform breakdown.
- `GET /api/analytics/heatmap`: returns posted counts for days with activity in the last 90 days.

Activity:
- `GET /api/activity`: query `limit`, `offset`, optional `video_id`; returns paginated activity entries with video title.

n8n:
- `GET /api/n8n/queue`: secret-protected; returns due scheduled videos.
- `POST /api/n8n/webhook/posted`: secret-protected; marks video posted and logs execution.
- `POST /api/n8n/webhook/failed`: secret-protected; logs failure only.

Health:
- `GET /api/health`: basic API health.
- `GET /api/health/n8n`: public reachability check for the sidebar dot.

## 9. Key Components Reference

- `DropZone`: drag/drop and file picker, per-file upload form, `TagInput`, `UploadProgress`; emits `upload-complete`.
- `VideoCard`: thumbnail/placeholder, hover controls, status badge, metadata, tags, bulk checkbox; opens `VideoModal`.
- `VideoGrid`: responsive grid, skeletons, empty state, card stagger.
- `VideoModal`: video playback, editable metadata, tags, status action, delete dialog, mini `ActivityFeed`.
- `BulkActionBar`: selected count, reschedule, draft, delete selected.
- `FilterBar`: debounced search, platform controls, date range, sort, tag chips, mobile sheet.
- `TagInput`: chip input, autocomplete, create-new option, `v-model` string array.
- `ActivityFeed`: paginated activity list, n8n chip, auto-refresh every 30 seconds.
- `StatCard`: KPI tile with optional trend and skeleton state.
- `PostingHeatmap`: pure CSS grid for last 90 days.

## 10. State Management Reference

`videoStore`: holds `videos`, `filters`, `selectedIds`, `loading`, `selectedVideo`, `total`. Actions fetch, upload, update, status-change, bulk action, delete, filter mutate, select, clear, open.

`tagStore`: holds `tags`; `fetchTags()` calls `GET /api/tags`.

`activityStore`: holds `entries`, `total`, `loading`, pagination. `fetchActivity()` resets page, `loadMore()` appends.

`uiStore`: holds `activeModal`, `toasts`, `isDraggingFile`. Toasts auto-dismiss after duration.

## 11. n8n Integration Guide

n8n should use an HTTP Request node to poll `GET /api/n8n/queue` with `x-n8n-secret`. For each returned video, the workflow can download or stream `/api/videos/:id/file`, post to a platform, then call either success or failure webhook. A success call marks `posted`; a failure call leaves status unchanged. Inspect `n8n_execution_log` on the video record and activity entries to debug.

## 12. FFmpeg & Thumbnail System

`thumbnailService.processVideo(inputPath, outputBasename)` probes duration and extracts one `640x360` JPEG frame near one second. The service catches FFmpeg errors and returns `null` fields instead of throwing, so uploads still succeed when video processing fails.

## 13. Tag System

Tags are normalized to avoid duplicated free-text metadata. `tagService.upsertTags(videoId, tagNames)` clears current links, inserts missing tag names with `INSERT OR IGNORE`, and writes junction rows. API responses call `attachTags()` so the frontend always receives plain `tags: string[]`.

## 14. Analytics Reference

- `totalUploaded`: count of all videos.
- `totalPosted`: count where `status = posted`.
- `totalScheduled`: count where `status = scheduled`.
- `totalDraft`: count where `status = draft`.
- `postsThisWeek`: posted rows in the current week window.
- `postsLastWeek`: posted rows in the previous week window.
- `weeklyTrend`: percentage change from last week.
- `avgPostsPerWeek`: posted count divided by distinct posted weeks.
- `mostActivePlatform`: posted platform with highest count.
- `totalStorageBytes`: sum of video file sizes.
- `platformBreakdown`: grouped counts by platform.
- `heatmap`: grouped posted count by `date(posted_at)`.

## 15. Activity Log Reference

Actions: `uploaded`, `edited`, `status_changed`, `deleted`, `n8n_queued`, `n8n_posted`, `n8n_failed`. Source is `user` for UI actions and `n8n` for automation events. Calls use `activityService.log(...).catch(console.error)` so logging never blocks the main operation.

## 16. Adding a New Platform

1. Add the platform in `client/src/constants/index.js`.
2. Add or reuse a Lucide icon in components that display platforms.
3. Add a CSS segment color if the analytics breakdown should show a distinct bar.
4. Update n8n workflow platform routing.
5. No migration is required because `platform` is a string.

## 17. Deployment Notes

Run `npm run build` to create `client/dist`. In production, serve that static directory from Express or a static host, run `NODE_ENV=production`, keep SQLite/uploads/thumbnails on persistent storage, and rotate `N8N_WEBHOOK_SECRET` when n8n credentials change. Back up `server/db/*.sqlite` and file storage together.

## 18. Known Limitations

- Single-user; no login or role model.
- No video re-encoding or transcoding.
- SQLite is excellent for this local v1 but not for high-concurrency multi-user workloads.
- Social platform posting is delegated to n8n; V.C.M does not call platform APIs directly.

## 19. Responsive Design Reference

Sidebar: fixed 240px on `lg+`, icon rail on `md`, drawer below `md`. Video grid: 4/3/2/1 columns from `xl` to mobile. Video modal: right panel on `lg+`, bottom sheet below. FilterBar: full row on `md+`, bottom sheet below. Analytics: 4-column KPI grid on `lg+`, 2-column KPI on smaller screens. Touch targets use at least 44px height, and cards support long-press context actions on touch.

## 20. Animation & Interaction Reference

Route transitions fade and translate by 8px using `--transition-base`. Cards stagger with 40ms intervals capped at 10. Card hover lifts thumbnail brightness and reveals overlay controls. Bulk bar uses spring slide-up. Toasts slide from the right and auto-dismiss. Filter chips scale/fade. Sidebar active indicator moves vertically. n8n status dot pulses when reachable. Skeleton loaders use a shimmer keyframe.
