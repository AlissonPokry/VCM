# V.C.M (Vertical Content Manager)

V.C.M is a local-first vertical video scheduling dashboard for Reels, TikTok, and YouTube Shorts, with n8n-ready automation endpoints.

<!-- Add screenshot here -->

## Quick Start

```bash
git clone <repo>
npm install
cp .env.example .env   # fill in values
npm run migrate
npm run dev
```

Backend requirements: Go 1.22+, GCC for CGO SQLite builds, and FFmpeg on `PATH` for thumbnails. If FFmpeg is missing, uploads still work but thumbnail extraction is disabled.

## Core Features

- Drag-and-drop vertical video upload with Go/FFmpeg thumbnail and duration extraction
- Scheduled, posted, and draft video queues with filters, search, tags, and bulk actions
- Editable video detail panel with playback, status changes, tags, and activity timeline
- Analytics dashboard with KPIs, heatmap, platform breakdown, tag cloud, and weekly chart
- Activity feed with automatic upload/edit/status/delete/n8n audit events
- Secret-protected n8n queue and webhook endpoints with parallel Go platform dispatch

Full developer reference: [DOCUMENTATION.md](DOCUMENTATION.md)  
Go backend refactor notes: [GO_BACKEND_REFACTOR.md](GO_BACKEND_REFACTOR.md)  
n8n setup guide: [N8N_INTEGRATION.md](N8N_INTEGRATION.md)
# VCM
# VCM
