const fs = require('fs/promises');
const path = require('path');
const db = require('../db/connection');
const tagService = require('./tagService');
const activityService = require('./activityService');
const thumbnailService = require('./thumbnailService');
const { AppError } = require('../middleware/errorHandler');
const { uploadDir } = require('../middleware/upload');

const VALID_STATUSES = ['scheduled', 'posted', 'draft'];
const VALID_PLATFORMS = ['instagram', 'tiktok', 'youtube', 'all'];

const parseTags = (tags) => {
  if (Array.isArray(tags)) return tags;
  if (!tags) return [];
  try {
    const parsed = JSON.parse(tags);
    return Array.isArray(parsed) ? parsed : tagService.normalizeTagNames(tags);
  } catch {
    const cleaned = String(tags).replace(/^\[/, '').replace(/\]$/, '').replace(/["']/g, '');
    return tagService.normalizeTagNames(cleaned);
  }
};

function ensureChoice(value, choices, field) {
  if (value && !choices.includes(value)) {
    throw new AppError(`Invalid ${field}`, 422, `INVALID_${field.toUpperCase()}`);
  }
}

async function getById(id) {
  const video = await db('videos').where({ id }).first();
  if (!video) {
    throw new AppError('Video not found', 404, 'VIDEO_NOT_FOUND');
  }
  return tagService.attachTags(video);
}

function applyFilters(query, filters = {}) {
  const {
    status,
    platform,
    tag,
    search,
    sort = 'created_at',
    order = 'desc',
    dateFrom,
    dateTo
  } = filters;

  if (status && status !== 'all') query.where('v.status', status);
  if (platform && platform !== 'all') query.where('v.platform', platform);
  if (search) {
    query.where((builder) => {
      builder.where('v.title', 'like', `%${search}%`).orWhere('v.description', 'like', `%${search}%`);
    });
  }
  if (dateFrom) query.where('v.scheduled_at', '>=', dateFrom);
  if (dateTo) query.where('v.scheduled_at', '<=', dateTo);

  const tags = Array.isArray(tag) ? tag : tag ? String(tag).split(',') : [];
  tags.filter(Boolean).forEach((tagName) => {
    query.whereExists(function tagExists() {
      this.select(db.raw('1'))
        .from('video_tags as ft')
        .join('tags as tt', 'ft.tag_id', 'tt.id')
        .whereRaw('ft.video_id = v.id')
        .whereRaw('tt.name = ? COLLATE NOCASE', [tagName.trim()]);
    });
  });

  const allowedSorts = ['created_at', 'scheduled_at', 'posted_at', 'title'];
  query.orderBy(`v.${allowedSorts.includes(sort) ? sort : 'created_at'}`, order === 'asc' ? 'asc' : 'desc');
}

async function list(filters = {}) {
  const base = db('videos as v').select('v.*');
  applyFilters(base, filters);
  const countQuery = db('videos as v').countDistinct({ total: 'v.id' });
  applyFilters(countQuery, filters);

  const [videos, totalRow] = await Promise.all([base, countQuery.first()]);
  return { videos: await tagService.attachTags(videos), total: Number(totalRow?.total || 0) };
}

async function createFromUpload(file, body) {
  if (!file) {
    throw new AppError('Video file is required', 422, 'FILE_REQUIRED');
  }

  ensureChoice(body.platform, VALID_PLATFORMS, 'platform');

  const title = body.title?.trim() || path.parse(file.originalname).name;
  const payload = {
    title,
    description: body.description || '',
    filename: file.filename,
    original_name: file.originalname,
    file_size: file.size,
    platform: body.platform || 'instagram',
    status: 'scheduled',
    scheduled_at: body.scheduled_at || null,
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString()
  };

  const ids = await db('videos').insert(payload);
  const id = ids[0];

  const { thumbnailPath, duration } = await thumbnailService.processVideo(
    path.join(uploadDir, file.filename),
    path.parse(file.filename).name
  );

  await db('videos').where({ id }).update({
    thumbnail: thumbnailPath,
    duration,
    updated_at: new Date().toISOString()
  });
  await tagService.upsertTags(id, parseTags(body.tags));

  activityService
    .log(id, 'uploaded', `Video "${title}" uploaded (${Math.round(file.size / 1024 / 1024)}MB, ${duration || 0}s)`, 'user')
    .catch(console.error);

  return getById(id);
}

async function update(id, data) {
  const current = await getById(id);
  ensureChoice(data.platform, VALID_PLATFORMS, 'platform');

  const updates = {};
  ['title', 'description', 'platform', 'scheduled_at', 'n8n_workflow_id'].forEach((key) => {
    if (Object.prototype.hasOwnProperty.call(data, key)) updates[key] = data[key] || null;
  });

  if (Object.keys(updates).length > 0) {
    updates.updated_at = new Date().toISOString();
    await db('videos').where({ id }).update(updates);
  }

  if (Object.prototype.hasOwnProperty.call(data, 'tags')) {
    await tagService.upsertTags(id, parseTags(data.tags));
  }

  activityService.log(id, 'edited', `Metadata updated for "${updates.title || current.title}"`, 'user').catch(console.error);
  return getById(id);
}

async function updateStatus(id, status) {
  ensureChoice(status, VALID_STATUSES, 'status');
  await getById(id);

  const updates = {
    status,
    updated_at: new Date().toISOString()
  };

  if (status === 'posted') updates.posted_at = new Date().toISOString();
  if (status !== 'posted') updates.posted_at = null;

  await db('videos').where({ id }).update(updates);
  activityService.log(id, 'status_changed', `Status changed to ${status}`, 'user').catch(console.error);
  return getById(id);
}

async function deleteVideo(id) {
  const video = await getById(id);
  await db('video_tags').where({ video_id: id }).del();
  await db('videos').where({ id }).del();

  await Promise.allSettled([
    fs.unlink(path.join(uploadDir, video.filename)),
    video.thumbnail ? fs.unlink(path.resolve(__dirname, '../../', video.thumbnail)) : Promise.resolve()
  ]);

  activityService.log(null, 'deleted', `Video "${video.title}" was permanently deleted`, 'user').catch(console.error);
  return { id: Number(id) };
}

async function bulkAction(ids = [], action, opts = {}) {
  const safeIds = ids.map(Number).filter(Boolean);
  if (safeIds.length === 0) {
    throw new AppError('At least one video id is required', 422, 'IDS_REQUIRED');
  }

  if (action === 'delete') {
    const results = [];
    for (const id of safeIds) {
      results.push(await deleteVideo(id));
    }
    return results;
  }

  if (action === 'draft') {
    await db('videos').whereIn('id', safeIds).update({ status: 'draft', posted_at: null, updated_at: new Date().toISOString() });
    safeIds.forEach((id) => activityService.log(id, 'status_changed', 'Status changed to draft', 'user').catch(console.error));
    return safeIds;
  }

  if (action === 'reschedule') {
    if (!opts.scheduled_at) {
      throw new AppError('scheduled_at is required for reschedule', 422, 'SCHEDULED_AT_REQUIRED');
    }
    await db('videos').whereIn('id', safeIds).update({ status: 'scheduled', scheduled_at: opts.scheduled_at, updated_at: new Date().toISOString() });
    safeIds.forEach((id) => activityService.log(id, 'status_changed', `Status changed to scheduled for ${opts.scheduled_at}`, 'user').catch(console.error));
    return safeIds;
  }

  throw new AppError('Invalid bulk action', 422, 'INVALID_BULK_ACTION');
}

async function getDueScheduled() {
  const videos = await db('videos as v')
    .where('v.status', 'scheduled')
    .whereNotNull('v.scheduled_at')
    .where('v.scheduled_at', '<=', new Date().toISOString())
    .select('v.*')
    .orderBy('v.scheduled_at', 'asc');

  const withTags = await tagService.attachTags(videos);
  withTags.forEach((video) => {
    activityService.log(video.id, 'n8n_queued', `Picked up by n8n for posting to ${video.platform}`, 'n8n').catch(console.error);
  });
  return withTags;
}

async function appendExecution(id, entry) {
  const video = await getById(id);
  let log = [];
  try {
    log = JSON.parse(video.n8n_execution_log || '[]');
  } catch {
    log = [];
  }

  log.push({
    execution_id: entry.execution_id || null,
    timestamp: new Date().toISOString(),
    result: entry.result,
    platform: entry.platform || null,
    error: entry.error || null
  });

  await db('videos').where({ id }).update({
    n8n_execution_log: JSON.stringify(log),
    updated_at: new Date().toISOString()
  });
}

async function markPostedFromN8n({ video_id: videoId, posted_at: postedAt, platform, execution_id: executionId }) {
  ensureChoice(platform, VALID_PLATFORMS, 'platform');
  await appendExecution(videoId, { execution_id: executionId, result: 'posted', platform, error: null });
  await db('videos').where({ id: videoId }).update({
    status: 'posted',
    posted_at: postedAt || new Date().toISOString(),
    platform,
    updated_at: new Date().toISOString()
  });
  activityService.log(videoId, 'n8n_posted', `Successfully posted to ${platform} at ${postedAt || 'now'}`, 'n8n').catch(console.error);
  return getById(videoId);
}

async function markFailedFromN8n({ video_id: videoId, error, execution_id: executionId }) {
  await appendExecution(videoId, { execution_id: executionId, result: 'failed', platform: null, error: error || 'Unknown error' });
  activityService.log(videoId, 'n8n_failed', `n8n execution failed: ${error || 'Unknown error'}`, 'n8n').catch(console.error);
  return getById(videoId);
}

module.exports = {
  list,
  getById,
  createFromUpload,
  update,
  updateStatus,
  deleteVideo,
  bulkAction,
  getDueScheduled,
  markPostedFromN8n,
  markFailedFromN8n
};
