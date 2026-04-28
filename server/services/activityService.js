const db = require('../db/connection');

async function log(videoId, action, detail, source = 'user') {
  await db('activity_log').insert({
    video_id: videoId || null,
    action,
    detail,
    source,
    created_at: new Date().toISOString()
  });
}

async function listActivity({ limit = 50, offset = 0, video_id: videoId } = {}) {
  const safeLimit = Math.min(Math.max(Number(limit) || 50, 1), 100);
  const safeOffset = Math.max(Number(offset) || 0, 0);

  const base = db('activity_log as a')
    .leftJoin('videos as v', 'a.video_id', 'v.id')
    .select('a.*', 'v.title as video_title')
    .orderBy('a.created_at', 'desc');

  if (videoId) {
    base.where('a.video_id', videoId);
  }

  const totalRow = await base.clone().clearSelect().clearOrder().count({ total: 'a.id' }).first();
  const entries = await base.limit(safeLimit).offset(safeOffset);
  return { entries, total: Number(totalRow?.total || 0) };
}

module.exports = { log, listActivity };
