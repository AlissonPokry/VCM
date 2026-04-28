export function formatDate(value) {
  if (!value) return 'Unscheduled';
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value));
}

export function formatDateInput(value) {
  if (!value) return '';
  const date = new Date(value);
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

export function toIsoFromInput(value) {
  if (!value) return null;
  return new Date(value).toISOString();
}

export function formatDuration(seconds) {
  if (!seconds) return '0:00';
  const mins = Math.floor(seconds / 60);
  const secs = Math.round(seconds % 60).toString().padStart(2, '0');
  return `${mins}:${secs}`;
}

export function formatFileSize(bytes) {
  if (!bytes) return '0 MB';
  const units = ['B', 'KB', 'MB', 'GB'];
  let size = bytes;
  let unit = 0;
  while (size >= 1024 && unit < units.length - 1) {
    size /= 1024;
    unit += 1;
  }
  return `${size.toFixed(unit === 0 ? 0 : 1)} ${units[unit]}`;
}

export function relativeTime(value) {
  if (!value) return '';
  const diff = Math.round((new Date(value).getTime() - Date.now()) / 1000);
  const formatter = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
  const units = [
    ['year', 31536000],
    ['month', 2592000],
    ['week', 604800],
    ['day', 86400],
    ['hour', 3600],
    ['minute', 60],
    ['second', 1]
  ];
  const [unit, seconds] = units.find(([, amount]) => Math.abs(diff) >= amount) || ['second', 1];
  return formatter.format(Math.round(diff / seconds), unit);
}

export function lastNDays(days) {
  return Array.from({ length: days }, (_, index) => {
    const date = new Date();
    date.setDate(date.getDate() - (days - 1 - index));
    return date.toISOString().slice(0, 10);
  });
}

export function groupByWeek(heatmapRows = [], weeks = 12) {
  const counts = new Map(heatmapRows.map((row) => [row.date, Number(row.count || 0)]));
  const days = lastNDays(weeks * 7);
  const grouped = [];

  for (let i = 0; i < weeks; i += 1) {
    const slice = days.slice(i * 7, i * 7 + 7);
    grouped.push({
      label: slice[0].slice(5),
      count: slice.reduce((total, date) => total + (counts.get(date) || 0), 0)
    });
  }

  return grouped;
}
