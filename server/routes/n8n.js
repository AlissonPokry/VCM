const express = require('express');
const videoService = require('../services/videoService');
const { AppError } = require('../middleware/errorHandler');

/**
 * V.C.M is the source of truth. n8n polls due scheduled videos, posts them,
 * then reports success/failure through secret-protected webhooks. Failed posts
 * do not change video status so humans can review or retry.
 */
const router = express.Router();

function requireSecret(req, res, next) {
  if (!process.env.N8N_WEBHOOK_SECRET || req.headers['x-n8n-secret'] !== process.env.N8N_WEBHOOK_SECRET) {
    return next(new AppError('Invalid n8n secret', 401, 'INVALID_N8N_SECRET'));
  }
  return next();
}

router.use(requireSecret);

router.get('/queue', async (req, res, next) => {
  try {
    const videos = await videoService.getDueScheduled();
    res.json({ success: true, data: videos, meta: { total: videos.length } });
  } catch (error) {
    next(error);
  }
});

router.post('/webhook/posted', async (req, res, next) => {
  try {
    const video = await videoService.markPostedFromN8n(req.body);
    res.json({ success: true, data: video, meta: {} });
  } catch (error) {
    next(error);
  }
});

router.post('/webhook/failed', async (req, res, next) => {
  try {
    const video = await videoService.markFailedFromN8n(req.body);
    res.json({ success: true, data: video, meta: {} });
  } catch (error) {
    next(error);
  }
});

module.exports = router;
