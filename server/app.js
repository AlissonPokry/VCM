const express = require('express');
const cors = require('cors');
const path = require('path');

require('dotenv').config({ path: path.resolve(__dirname, '../.env') });

const videos = require('./routes/videos');
const tags = require('./routes/tags');
const analytics = require('./routes/analytics');
const activity = require('./routes/activity');
const n8n = require('./routes/n8n');
const health = require('./routes/health');
const { errorHandler } = require('./middleware/errorHandler');

const app = express();

app.use(cors({ origin: process.env.CORS_ORIGIN || 'http://localhost:5173' }));
app.use(express.json({ limit: '10mb' }));
app.use(express.urlencoded({ extended: true }));

app.get('/api/health', (req, res) => {
  res.json({ success: true, data: { status: 'ok' }, meta: {} });
});

app.use('/api/health', health);
app.use('/api/videos', videos);
app.use('/api/tags', tags);
app.use('/api/analytics', analytics);
app.use('/api/activity', activity);
app.use('/api/n8n', n8n);

app.use(errorHandler);

module.exports = app;
