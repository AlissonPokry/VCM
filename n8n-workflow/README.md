# 📡 Reel Queue — n8n Social Media Posting Workflow

Automated workflow that polls Reel Queue for due videos, downloads from Google Drive, and posts to YouTube + TikTok.

---

## 🏗️ Architecture

```
┌──────────────┐  GET /api/n8n/queue     ┌─────────────────────────────────────┐
│  Reel Queue  │ ◄────────────────────── │          n8n Workflow               │
│  (Express)   │                         │                                     │
│              │  POST /webhook/posted   │  Schedule Trigger (every 5 min)     │
│              │ ◄────────────────────── │    ↓                                │
│              │  POST /webhook/failed   │  Fetch Queue → IF has videos        │
│              │ ◄────────────────────── │    ↓                                │
└──────────────┘                         │  Loop → Download → Route → Upload   │
                                         │    ↓                                │
┌──────────────┐                         │  Report success/failure             │
│ Google Drive │ ◄── download ────────── │                                     │
│ YouTube*     │ ◄── upload ──────────── │  * = placeholder until API ready    │
│ TikTok*      │ ◄── upload ──────────── │                                     │
└──────────────┘                         └─────────────────────────────────────┘
```

---

## 📂 Files

| File | Purpose |
|------|---------|
| `workflow.json` | Main workflow — import into n8n |
| `credentials-setup.md` | Step-by-step credential configuration |
| `env-template.txt` | n8n environment variables needed |
| `docker-compose.yml` | Docker Compose to run n8n locally |
| `README.md` | This file |

---

## 🐳 Docker Quick Start

```bash
cd n8n-workflow

# Start n8n
docker-compose up -d

# Open n8n UI
# → http://localhost:5678
# Login: admin / change_me (change in docker-compose.yml)
```

**Important:** `REEL_QUEUE_API_URL` uses `host.docker.internal` so n8n container can reach Reel Queue running on your host machine at port 3001.

### Import workflow after starting:
1. Open http://localhost:5678
2. **Workflows** → **Import from File**
3. Select `workflow.json` from this folder
4. Configure credentials (see `credentials-setup.md`)
5. Activate workflow

### Stop n8n:
```bash
docker-compose down
```

### Reset data:
```bash
docker-compose down -v   # removes volume (all n8n data lost)
```

---

## 🚀 Quick Start (Manual — no Docker)

### 1. Import Workflow
1. Open n8n instance
2. **Workflows** → **Import from File**
3. Select `workflow.json`

### 2. Set Environment Variables
Add to n8n environment (`.env` or n8n Settings):
```env
REEL_QUEUE_API_URL=http://localhost:3001
N8N_WEBHOOK_SECRET=change_me_before_production
```

### 3. Configure Credentials
See `credentials-setup.md` for detailed steps.

**Required now:**
- Google Drive OAuth2 (for downloading videos)
- HTTP Header Auth (for Reel Queue `x-n8n-secret`)

**Required later (placeholder nodes):**
- YouTube OAuth2 (when YouTube Data API v3 ready)
- TikTok API (when TikTok Developer App ready)

### 4. Activate
Enable the workflow — it polls every 5 minutes for due videos.

---

## 🔄 Workflow Flow

```
1. Schedule Trigger fires every 5 min
2. HTTP Request → GET /api/n8n/queue (with x-n8n-secret header)
3. IF node checks if any videos are due
4. SplitInBatches processes one video at a time
5. Set node extracts: videoId, title, description, platform, googleDriveFileId, tags
6. Google Drive downloads video file by ID
7. Switch routes by platform:
   ├─ "youtube" → Upload to YouTube* → Report Success
   ├─ "tiktok"  → Upload to TikTok*  → Report Success
   └─ "all"     → Both uploads*      → Merge → Report Success
8. Success: POST /api/n8n/webhook/posted → Reel Queue marks video as 'posted'
9. Error:   POST /api/n8n/webhook/failed → Reel Queue logs error, status unchanged
```

---

## 📋 Node Reference

| # | Node | Type | Notes |
|---|------|------|-------|
| 1 | Poll Every 5 Min | Schedule Trigger | Configurable interval |
| 2 | Fetch Due Videos | HTTP Request | GET queue endpoint |
| 3 | Has Due Videos? | IF | Checks array not empty |
| 4 | No Videos Due | NoOp | End path |
| 5 | Process Each Video | SplitInBatches | Batch size: 1 |
| 6 | Extract Video Data | Set | Maps fields |
| 7 | Download from Google Drive | Google Drive | OAuth2 required |
| 8 | Route by Platform | Switch | youtube/tiktok/all |
| 9 | Upload to YouTube | HTTP Request | ⚠️ Placeholder |
| 10 | Upload to TikTok | HTTP Request | ⚠️ Placeholder |
| 11 | YouTube (All Platforms) | HTTP Request | ⚠️ Placeholder |
| 12 | TikTok (All Platforms) | HTTP Request | ⚠️ Placeholder |
| 13 | Merge All Results | Merge | Combines parallel results |
| 14-16 | Report Success (×3) | HTTP Request | POST /webhook/posted |
| 17-18 | Next Video (×2) | NoOp | Loop connectors |
| 19 | On Error | Error Trigger | Catches failures |
| 20 | Report Failure | HTTP Request | POST /webhook/failed |

---

## ⚠️ Placeholder Nodes

YouTube and TikTok upload nodes are **placeholders** — they won't work until you:

1. **YouTube**: Set up Google Cloud Project + YouTube Data API v3 + OAuth2
2. **TikTok**: Register TikTok Developer App + Content Posting API access

Each placeholder node has detailed setup instructions in its **Notes** field within n8n.

---

## 🔧 Troubleshooting

| Problem | Fix |
|---------|-----|
| Queue returns empty | Check `scheduled_at` is in the past, `status` is `scheduled` |
| 401 on queue fetch | Verify `N8N_WEBHOOK_SECRET` matches Reel Queue `.env` |
| Google Drive download fails | Check OAuth2 credentials, file ID valid |
| Video stuck in `scheduled` | Check n8n execution log, verify webhook secret |
| Error trigger not firing | Ensure workflow error handling is enabled in settings |
