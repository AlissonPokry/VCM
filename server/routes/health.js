const express = require('express');

const router = express.Router();

router.get('/n8n', (req, res) => {
  res.json({
    success: true,
    data: {
      reachable: true,
      protected: true
    },
    meta: {}
  });
});

module.exports = router;
