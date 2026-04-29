export const PLATFORMS = [
  { value: 'all', label: 'All', color: '#6c63ff', icon: 'Zap' },
  { value: 'instagram', label: 'Instagram', color: '#e879f9', icon: 'Instagram' },
  { value: 'tiktok', label: 'TikTok', color: '#22d3a5', icon: 'Music2' },
  { value: 'youtube', label: 'YouTube', color: '#ff6b6b', icon: 'Youtube' }
];

export const STATUSES = [
  { value: 'scheduled', label: 'Scheduled', classes: 'text-amber-400 bg-amber-400/10' },
  { value: 'partial', label: 'Partial', classes: 'text-orange-400 bg-orange-400/10' },
  { value: 'posted', label: 'Posted', classes: 'text-emerald-400 bg-emerald-400/10' },
  { value: 'draft', label: 'Draft', classes: 'text-muted bg-white/5' }
];

export const SORT_OPTIONS = [
  { value: 'created_at', label: 'Created' },
  { value: 'scheduled_at', label: 'Schedule' },
  { value: 'posted_at', label: 'Posted' },
  { value: 'title', label: 'Title' }
];

export const ACTIVITY_ACTIONS = {
  uploaded: { label: 'Uploaded', color: 'text-accent', icon: 'Upload' },
  edited: { label: 'Edited', color: 'text-secondary', icon: 'FilePen' },
  status_changed: { label: 'Status', color: 'text-amber', icon: 'CheckCircle2' },
  deleted: { label: 'Deleted', color: 'text-warm', icon: 'Trash2' },
  n8n_queued: { label: 'Queued', color: 'text-accent', icon: 'Zap' },
  n8n_posted: { label: 'Posted', color: 'text-green', icon: 'Send' },
  n8n_failed: { label: 'Failed', color: 'text-warm', icon: 'AlertTriangle' }
};
