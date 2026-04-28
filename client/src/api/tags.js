import { api } from './videos';

export async function getTags() {
  const response = await api.get('/api/tags');
  return response.data;
}
