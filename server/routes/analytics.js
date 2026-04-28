const express = require('express');
const analyticsService = require('../services/analyticsService');

const router = express.Router();

router.get('/summary', async (req, res, next) => {
  try {
    res.json({ success: true, data: await analyticsService.summary(), meta: {} });
  } catch (error) {
    next(error);
  }
});

router.get('/heatmap', async (req, res, next) => {
  try {
    const data = await analyticsService.heatmap();
    res.json({ success: true, data, meta: { total: data.length } });
  } catch (error) {
    next(error);
  }
});

module.exports = router;
