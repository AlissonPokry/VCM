#!/bin/sh
# Auto-setup n8n: create owner account + import workflow on first startup
# Uses Node.js for HTTP calls (guaranteed available in n8n container)

MARKER_FILE="/home/node/.n8n/.setup_complete"
WORKFLOW_FILE="/home/node/workflow.json"

# Owner account credentials (from environment or defaults)
OWNER_EMAIL="${N8N_OWNER_EMAIL:-admin@reelqueue.local}"
OWNER_FIRST="${N8N_OWNER_FIRSTNAME:-Admin}"
OWNER_LAST="${N8N_OWNER_LASTNAME:-ReelQueue}"
OWNER_PASS="${N8N_OWNER_PASSWORD:-ReelQueue2025!}"

if [ -f "$MARKER_FILE" ]; then
  echo "📋 Setup already complete. Starting n8n..."
  exec n8n start
fi

echo "🔄 First run — running workflow import via CLI..."

# Import workflow first (before n8n starts, CLI handles DB directly)
n8n import:workflow --input="$WORKFLOW_FILE" 2>&1
IMPORT_EXIT=$?

if [ $IMPORT_EXIT -eq 0 ]; then
  echo "✅ Workflow imported via CLI"
else
  echo "⚠️ CLI workflow import failed (exit: $IMPORT_EXIT)"
  echo "   You can import manually via n8n UI after startup"
fi

# Mark import done so we don't retry
touch "$MARKER_FILE"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  🎬 Reel Queue n8n starting..."
echo ""
echo "  URL:      http://localhost:5678"
echo ""
echo "  First login: create owner account in UI"
echo "  Suggested credentials:"
echo "    Email:    ${OWNER_EMAIL}"
echo "    Password: ${OWNER_PASS}"
echo ""
echo "  Workflow: should appear in Workflows list"
echo "  Next: configure Google Drive credentials"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Start n8n (runs migrations + starts server)
exec n8n start
