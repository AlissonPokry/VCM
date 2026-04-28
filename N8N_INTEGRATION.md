# V.C.M n8n Integration

## 1. Architecture Overview

```text
+------------------------------+     GET /api/n8n/queue      +-------------+
| V.C.M (Vertical Content Mgr)  | <-------------------------- |     n8n     |
| Express API + SQLite          |                             |  polling    |
|                              | ---- POST /webhook/posted -> |             |
|                              | ---- POST /webhook/failed -> |             |
+------------------------------+                             +-------------+
```

V.C.M owns the queue, files, metadata, status, and execution log. n8n polls due scheduled videos, posts them to social platforms, then calls back with success or failure.

## 2. Authentication

- Header name: `x-n8n-secret`
- Value: must match `N8N_WEBHOOK_SECRET` in `.env`
- Required on all `/api/n8n/*` routes
- Missing or incorrect values return `401` with `INVALID_N8N_SECRET`

## 3. Polling Trigger Setup

- Node type: HTTP Request
- URL: `http://your-server:3001/api/n8n/queue`
- Method: `GET`
- Header: `x-n8n-secret: your_secret`
- Suggested interval: every 5 minutes

The queue returns scheduled videos where `scheduled_at <= now`. Each returned video includes `tags`.

## 4. Success Webhook Payload

```json
POST /api/n8n/webhook/posted
Headers: { "x-n8n-secret": "your_secret", "Content-Type": "application/json" }
Body: {
  "video_id": 42,
  "posted_at": "2025-06-01T14:30:00Z",
  "platform": "instagram",
  "execution_id": "n8n-exec-abc123"
}
```

Success changes the video to `posted`, writes `posted_at`, appends one `n8n_execution_log` entry, and writes `n8n_posted` activity.

## 5. Failure Webhook Payload

```json
POST /api/n8n/webhook/failed
Headers: { "x-n8n-secret": "your_secret", "Content-Type": "application/json" }
Body: {
  "video_id": 42,
  "error": "Instagram API rate limit exceeded",
  "execution_id": "n8n-exec-abc123"
}
```

Failure appends an execution log entry and writes `n8n_failed` activity. It does not change video status, so the item remains available for review or retry.

## 6. Troubleshooting

- Video stuck in `scheduled`: check `scheduled_at`, video `status`, n8n execution history, then the video detail panel execution log.
- `401` on queue or webhooks: verify the n8n header exactly matches `N8N_WEBHOOK_SECRET`.
- Queue returns empty: ensure at least one video has `status = scheduled` and `scheduled_at` in the past.
- Sidebar n8n dot is green but queue returns `401`: the dot uses public `/api/health/n8n` for reachability only; n8n still needs the secret.
