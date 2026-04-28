import { api } from './videos';

export async function getActivity(params) {
  const response = await api.get('/api/activity', { params });
  return response.data;
}
