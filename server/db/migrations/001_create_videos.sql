CREATE TABLE IF NOT EXISTS videos (
  id               INTEGER PRIMARY KEY AUTOINCREMENT,
  title            TEXT NOT NULL,
  description      TEXT,
  filename         TEXT NOT NULL UNIQUE,
  original_name    TEXT NOT NULL,
  file_size        INTEGER NOT NULL,
  duration         INTEGER,
  thumbnail        TEXT,
  platform         TEXT DEFAULT 'instagram',
  status           TEXT DEFAULT 'scheduled',
  scheduled_at     DATETIME,
  posted_at        DATETIME,
  n8n_workflow_id  TEXT,
  n8n_execution_log TEXT DEFAULT '[]',
  created_at       DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at       DATETIME DEFAULT CURRENT_TIMESTAMP
);
