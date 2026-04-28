const fs = require('fs');
const path = require('path');
const knex = require('knex');

require('dotenv').config({ path: path.resolve(__dirname, '../../.env') });

const resolveProjectPath = (value, fallback) => {
  const target = value || fallback;
  return path.isAbsolute(target) ? target : path.resolve(__dirname, '../../', target);
};

const dbPath = resolveProjectPath(process.env.DB_PATH, './server/db/reel_queue.sqlite');
fs.mkdirSync(path.dirname(dbPath), { recursive: true });

const db = knex({
  client: 'better-sqlite3',
  connection: {
    filename: dbPath
  },
  pool: {
    afterCreate: (connection, done) => {
      // SQLite does not enforce FK constraints unless this pragma is set per connection.
      connection.pragma('foreign_keys = ON');
      done(null, connection);
    }
  },
  useNullAsDefault: true,
  migrations: {
    directory: path.resolve(__dirname, './migrations'),
    extension: 'js'
  }
});

module.exports = db;
