import axios from 'axios';
import { useUiStore } from '@/stores/uiStore';

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:3001'
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    const message = error.response?.data?.message || error.message || 'Request failed';
    try {
      useUiStore().pushToast({ message, type: 'error' });
    } catch {
      console.error(message);
    }
    return Promise.reject(error);
  }
);

const normalizeFilters = (filters = {}) => ({
  ...filters,
  tag: filters.tags?.length ? filters.tags.join(',') : undefined,
  tags: undefined
});

export async function getVideos(params) {
  const response = await api.get('/api/videos', { params: normalizeFilters(params) });
  return response.data;
}

export async function getVideo(id) {
  const response = await api.get(`/api/videos/${id}`);
  return response.data;
}

export async function uploadVideo(formData, onProgress) {
  const response = await api.post('/api/videos/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress
  });
  return response.data;
}

export async function updateVideo(id, data) {
  const response = await api.patch(`/api/videos/${id}`, data);
  return response.data;
}

export async function updateVideoStatus(id, status) {
  const response = await api.patch(`/api/videos/${id}/status`, { status });
  return response.data;
}

export async function bulkVideos(payload) {
  const response = await api.post('/api/videos/bulk', payload);
  return response.data;
}

export async function deleteVideo(id) {
  const response = await api.delete(`/api/videos/${id}`);
  return response.data;
}

export function thumbnailUrl(id) {
  return `${api.defaults.baseURL}/api/videos/${id}/thumbnail`;
}

export function videoFileUrl(id) {
  return `${api.defaults.baseURL}/api/videos/${id}/file`;
}
