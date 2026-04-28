const express = require('express');
const path = require('path');
const fs = require('fs');
const videoService = require('../services/videoService');
const { upload, uploadDir } = require('../middleware/upload');
const { AppError } = require('../middleware/errorHandler');

const router = express.Router();

const sendSuccess = (res, data, meta = {}) => res.json({ success: true, data, meta });

router.get('/', async (req, res, next) => {
  try {
    const { videos, total } = await videoService.list(req.query);
    sendSuccess(res, videos, { total });
  } catch (error) {
    next(error);
  }
});

router.get('/:id', async (req, res, next) => {
  try {
    sendSuccess(res, await videoService.getById(req.params.id));
  } catch (error) {
    next(error);
  }
});

router.post('/upload', upload.single('file'), async (req, res, next) => {
  try {
    const video = await videoService.createFromUpload(req.file, req.body);
    sendSuccess(res.status(201), video, { total: 1 });
  } catch (error) {
    next(error);
  }
});

router.patch('/:id', async (req, res, next) => {
  try {
    sendSuccess(res, await videoService.update(req.params.id, req.body));
  } catch (error) {
    next(error);
  }
});

router.patch('/:id/status', async (req, res, next) => {
  try {
    sendSuccess(res, await videoService.updateStatus(req.params.id, req.body.status));
  } catch (error) {
    next(error);
  }
});

router.post('/bulk', async (req, res, next) => {
  try {
    sendSuccess(res, await videoService.bulkAction(req.body.ids, req.body.action, req.body));
  } catch (error) {
    next(error);
  }
});

router.delete('/:id', async (req, res, next) => {
  try {
    sendSuccess(res, await videoService.deleteVideo(req.params.id));
  } catch (error) {
    next(error);
  }
});

router.get('/:id/thumbnail', async (req, res, next) => {
  try {
    const video = await videoService.getById(req.params.id);
    if (!video.thumbnail) throw new AppError('Thumbnail not found', 404, 'THUMBNAIL_NOT_FOUND');
    const thumbnailPath = path.resolve(__dirname, '../../', video.thumbnail);
    if (!fs.existsSync(thumbnailPath)) throw new AppError('Thumbnail not found', 404, 'THUMBNAIL_NOT_FOUND');
    res.sendFile(thumbnailPath);
  } catch (error) {
    next(error);
  }
});

router.get('/:id/file', async (req, res, next) => {
  try {
    const video = await videoService.getById(req.params.id);
    const filePath = path.join(uploadDir, video.filename);
    if (!fs.existsSync(filePath)) throw new AppError('Video file not found', 404, 'VIDEO_FILE_NOT_FOUND');
    res.setHeader('Content-Type', 'video/mp4');
    res.sendFile(filePath);
  } catch (error) {
    next(error);
  }
});

module.exports = router;
