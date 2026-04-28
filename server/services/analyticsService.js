const db = require('../db/connection');

async function summary() {
  const row = await db.raw(`
    WITH
      totals AS (
        SELECT
          COUNT(*) AS totalUploaded,
          SUM(CASE WHEN status = 'posted' THEN 1 ELSE 0 END) AS totalPosted,
          SUM(CASE WHEN status = 'scheduled' THEN 1 ELSE 0 END) AS totalScheduled,
          SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END) AS totalDraft,
          COALESCE(SUM(file_size), 0) AS totalStorageBytes
        FROM videos
      ),
      weekly AS (
        SELECT
          SUM(CASE WHEN status = 'posted' AND posted_at >= strftime('%Y-%m-%d', 'now', 'weekday 0', '-7 days') THEN 1 ELSE 0 END) AS postsThisWeek,
          SUM(CASE WHEN status = 'posted' AND posted_at >= strftime('%Y-%m-%d', 'now', 'weekday 0', '-14 days') AND posted_at < strftime('%Y-%m-%d', 'now', 'weekday 0', '-7 days') THEN 1 ELSE 0 END) AS postsLastWeek
        FROM videos
      ),
      avg_week AS (
        SELECT
          CASE
            WHEN COUNT(DISTINCT strftime('%Y-%W', posted_at)) = 0 THEN 0
            ELSE ROUND(COUNT(*) * 1.0 / COUNT(DISTINCT strftime('%Y-%W', posted_at)), 1)
          END AS avgPostsPerWeek
        FROM videos
        WHERE status = 'posted' AND posted_at IS NOT NULL
      ),
      platform_rank AS (
        SELECT platform AS mostActivePlatform
        FROM videos
        WHERE status = 'posted'
        GROUP BY platform
        ORDER BY COUNT(*) DESC, platform ASC
        LIMIT 1
      )
    SELECT *
    FROM totals, weekly, avg_week
    LEFT JOIN platform_rank
  `);

  const data = row[0] || {};
  const platformRows = await db('videos')
    .select('platform')
    .count({ count: 'id' })
    .groupBy('platform');

  const postsThisWeek = Number(data.postsThisWeek || 0);
  const postsLastWeek = Number(data.postsLastWeek || 0);
  const trendBase = postsLastWeek === 0 ? (postsThisWeek > 0 ? 100 : 0) : Math.round(((postsThisWeek - postsLastWeek) / postsLastWeek) * 100);

  const platformBreakdown = { instagram: 0, tiktok: 0, youtube: 0, all: 0 };
  platformRows.forEach((item) => {
    platformBreakdown[item.platform] = Number(item.count || 0);
  });

  return {
    totalUploaded: Number(data.totalUploaded || 0),
    totalPosted: Number(data.totalPosted || 0),
    totalScheduled: Number(data.totalScheduled || 0),
    totalDraft: Number(data.totalDraft || 0),
    postsThisWeek,
    postsLastWeek,
    weeklyTrend: `${trendBase >= 0 ? '+' : ''}${trendBase}%`,
    avgPostsPerWeek: Number(data.avgPostsPerWeek || 0),
    mostActivePlatform: data.mostActivePlatform || null,
    totalStorageBytes: Number(data.totalStorageBytes || 0),
    platformBreakdown
  };
}

async function heatmap() {
  return db('videos')
    .select(db.raw("date(posted_at) as date"))
    .count({ count: 'id' })
    .where('status', 'posted')
    .whereNotNull('posted_at')
    .where('posted_at', '>=', db.raw("date('now', '-90 days')"))
    .groupByRaw('date(posted_at)')
    .orderBy('date', 'asc');
}

module.exports = { summary, heatmap };
