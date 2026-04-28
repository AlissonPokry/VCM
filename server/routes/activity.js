const express = require('express');
const activityService = require('../services/activityService');

const router = express.Router();

router.get('/', async (req, res, next) => {
  try {
    const { entries, total } = await activityService.listActivity(req.query);
    res.json({ success: true, data: entries, meta: { total } });
  } catch (error) {
    next(error);
  }
});

module.exports = router;
