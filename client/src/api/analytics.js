import { api } from './videos';

export async function getSummary() {
  const response = await api.get('/api/analytics/summary');
  return response.data;
}

export async function getHeatmap() {
  const response = await api.get('/api/analytics/heatmap');
  return response.data;
}
