CREATE TABLE IF NOT EXISTS activity_log (
  id         INTEGER PRIMARY KEY AUTOINCREMENT,
  video_id   INTEGER REFERENCES videos(id) ON DELETE SET NULL,
  action     TEXT NOT NULL,
  detail     TEXT,
  source     TEXT DEFAULT 'user',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_act_video   ON activity_log (video_id);
CREATE INDEX IF NOT EXISTS idx_act_created ON activity_log (created_at DESC);
