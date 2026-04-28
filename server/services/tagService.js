const db = require('../db/connection');

function normalizeTagNames(tagNames = []) {
  if (typeof tagNames === 'string') {
    return tagNames.split(',');
  }

  return tagNames
    .filter(Boolean)
    .map((name) => String(name).trim())
    .filter(Boolean)
    .filter((name, index, arr) => arr.findIndex((item) => item.toLowerCase() === name.toLowerCase()) === index);
}

async function upsertTags(videoId, tagNames = []) {
  const normalized = normalizeTagNames(tagNames);

  await db('video_tags').where({ video_id: videoId }).del();

  for (const name of normalized) {
    await db.raw('INSERT OR IGNORE INTO tags (name, created_at, updated_at) VALUES (?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)', [name]);
    const tag = await db('tags').whereRaw('name = ? COLLATE NOCASE', [name]).first();
    if (tag) {
      await db.raw('INSERT OR IGNORE INTO video_tags (video_id, tag_id) VALUES (?, ?)', [videoId, tag.id]);
    }
  }
}

async function attachTags(videos) {
  const list = Array.isArray(videos) ? videos : [videos];
  if (list.length === 0) return videos;

  const ids = list.map((video) => video.id);
  const rows = await db('video_tags as vt')
    .join('tags as t', 'vt.tag_id', 't.id')
    .whereIn('vt.video_id', ids)
    .select('vt.video_id', 't.name')
    .orderBy('t.name');

  const byVideo = rows.reduce((acc, row) => {
    acc[row.video_id] ||= [];
    acc[row.video_id].push(row.name);
    return acc;
  }, {});

  const withTags = list.map((video) => ({
    ...video,
    tags: byVideo[video.id] || []
  }));

  return Array.isArray(videos) ? withTags : withTags[0];
}

async function listTags() {
  return db('tags as t')
    .leftJoin('video_tags as vt', 't.id', 'vt.tag_id')
    .leftJoin('videos as v', 'vt.video_id', 'v.id')
    .select('t.id', 't.name')
    .count({ count: 'v.id' })
    .groupBy('t.id')
    .havingRaw('COUNT(v.id) > 0')
    .orderBy('count', 'desc')
    .orderBy('t.name', 'asc');
}

module.exports = { normalizeTagNames, upsertTags, attachTags, listTags };
