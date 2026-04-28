const express = require('express');
const tagService = require('../services/tagService');

const router = express.Router();

router.get('/', async (req, res, next) => {
  try {
    const tags = await tagService.listTags();
    res.json({ success: true, data: tags, meta: { total: tags.length } });
  } catch (error) {
    next(error);
  }
});

module.exports = router;
