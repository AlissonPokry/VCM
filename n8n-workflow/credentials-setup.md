# 🔑 Credential Setup Guide

Step-by-step instructions for configuring all credentials needed by the Reel Queue n8n workflow.

---

## 1. HTTP Header Auth (Reel Queue Secret) — REQUIRED

Used by: Fetch Due Videos, Report Success, Report Failure nodes.

### Setup in n8n:
1. Go to **Settings** → **Credentials** → **Add Credential**
2. Search for **Header Auth**
3. Configure:
   - **Name**: `Reel Queue Secret`
   - **Header Name**: `x-n8n-secret`
   - **Header Value**: Same value as `N8N_WEBHOOK_SECRET` in your Reel Queue `.env`

### Assign to nodes:
After creating, edit each HTTP Request node that calls Reel Queue and select this credential.

---

## 2. Google Drive OAuth2 — REQUIRED

Used by: Download from Google Drive node.

### Prerequisites:
- Google Cloud account
- A Google Cloud Project

### Setup:
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create or select a project
3. **APIs & Services** → **Library** → Enable **Google Drive API**
4. **APIs & Services** → **Credentials** → **Create Credentials** → **OAuth 2.0 Client ID**
   - Application type: **Web application**
   - Authorized redirect URIs: `https://your-n8n-domain/rest/oauth2-credential/callback`
   - For local: `http://localhost:5678/rest/oauth2-credential/callback`
5. Copy **Client ID** and **Client Secret**

### Setup in n8n:
1. Go to **Settings** → **Credentials** → **Add Credential**
2. Search for **Google Drive OAuth2 API**
3. Paste Client ID and Client Secret
4. Click **Connect my account** → Google login flow
5. Grant access to Google Drive

---

## 3. YouTube OAuth2 — LATER (when API ready)

Used by: Upload to YouTube placeholder nodes.

### Prerequisites:
- Same Google Cloud Project from step 2
- YouTube Data API v3 enabled

### Setup:
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. **APIs & Services** → **Library** → Enable **YouTube Data API v3**
3. Use the same OAuth2 Client ID from step 2, or create a new one
4. Add scope: `https://www.googleapis.com/auth/youtube.upload`

### Setup in n8n:
1. Go to **Settings** → **Credentials** → **Add Credential**
2. Search for **Google OAuth2 API** (or YouTube-specific if available)
3. Paste Client ID and Client Secret
4. Scopes: `https://www.googleapis.com/auth/youtube.upload`
5. Click **Connect my account**

### Replace placeholder node:
Option A: Replace HTTP Request with native **YouTube** node (if available in your n8n version)
Option B: Keep HTTP Request but update:
  - Authentication: OAuth2
  - Credential: Your YouTube OAuth2 credential
  - URL: `https://www.googleapis.com/upload/youtube/v3/videos?uploadType=resumable&part=snippet,status`
  - Body: Include snippet (title, description, tags) and status (privacyStatus)

---

## 4. TikTok API — LATER (when developer app ready)

Used by: Upload to TikTok placeholder nodes.

### Prerequisites:
- TikTok Developer account
- Approved TikTok app with `video.upload` and `video.publish` scopes

### Setup:
1. Go to [TikTok Developer Portal](https://developers.tiktok.com)
2. Create a new app
3. Request scopes:
   - `video.upload` — Upload video content
   - `video.publish` — Publish uploaded videos
4. Wait for approval (may take days)
5. Implement OAuth2 flow to get user access token

### Setup in n8n:
1. Go to **Settings** → **Credentials** → **Add Credential**
2. Search for **OAuth2 API** (generic)
3. Configure:
   - **Authorization URL**: `https://www.tiktok.com/v2/auth/authorize/`
   - **Access Token URL**: `https://open.tiktokapis.com/v2/oauth/token/`
   - **Client ID**: Your TikTok app's Client Key
   - **Client Secret**: Your TikTok app's Client Secret
   - **Scope**: `video.upload,video.publish`
4. Click **Connect my account**

### TikTok Upload Flow (2-step process):
```
Step 1: POST /v2/post/publish/video/init/
  → Returns: upload_url

Step 2: PUT {upload_url}
  → Send video binary
  → Returns: publish_id
```

### Replace placeholder node:
You'll need to split into 2 HTTP Request nodes:
1. **Init Upload** — POST to init endpoint, get upload_url
2. **Upload Binary** — PUT video data to upload_url

See TikTok docs: https://developers.tiktok.com/doc/content-posting-api-get-started

---

## Credential Summary

| Credential | Status | Used By |
|------------|--------|---------|
| HTTP Header Auth | ✅ Required now | Queue fetch, webhooks |
| Google Drive OAuth2 | ✅ Required now | Video download |
| YouTube OAuth2 | ⏳ Set up later | Video upload |
| TikTok OAuth2 | ⏳ Set up later | Video upload |
