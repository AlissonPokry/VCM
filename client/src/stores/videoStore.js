import { defineStore } from 'pinia';
import {
  bulkVideos,
  deleteVideo as deleteVideoApi,
  getVideo,
  getVideos,
  updateVideo as updateVideoApi,
  updateVideoStatus,
  uploadVideo as uploadVideoApi
} from '@/api/videos';
import { useTagStore } from './tagStore';
import { useUiStore } from './uiStore';

export const useVideoStore = defineStore('videos', {
  state: () => ({
    videos: [],
    filters: {
      status: 'all',
      platform: 'all',
      search: '',
      tags: [],
      dateFrom: null,
      dateTo: null,
      sort: 'created_at',
      order: 'desc'
    },
    selectedIds: [],
    loading: false,
    selectedVideo: null,
    total: 0
  }),
  getters: {
    hasSelection: (state) => state.selectedIds.length > 0
  },
  actions: {
    async fetchVideos(overrides = {}) {
      this.loading = true;
      try {
        const response = await getVideos({ ...this.filters, ...overrides });
        this.videos = response.data;
        this.total = response.meta?.total || response.data.length;
      } finally {
        this.loading = false;
      }
    },
    async fetchVideo(id) {
      const response = await getVideo(id);
      this.selectedVideo = response.data;
      return response.data;
    },
    async uploadVideo(formData, onProgress) {
      const response = await uploadVideoApi(formData, onProgress);
      await this.fetchVideos();
      await useTagStore().fetchTags();
      useUiStore().pushToast({ message: 'Video uploaded', type: 'success' });
      return response.data;
    },
    async updateVideo(id, data) {
      const response = await updateVideoApi(id, data);
      this.selectedVideo = response.data;
      await this.fetchVideos();
      await useTagStore().fetchTags();
      useUiStore().pushToast({ message: 'Video saved', type: 'success' });
      return response.data;
    },
    async updateStatus(id, status) {
      const response = await updateVideoStatus(id, status);
      this.selectedVideo = response.data;
      await this.fetchVideos();
      useUiStore().pushToast({ message: `Status changed to ${status}`, type: 'success' });
      return response.data;
    },
    async bulkAction(ids, action, opts = {}) {
      const response = await bulkVideos({ ids, action, ...opts });
      this.clearSelection();
      await this.fetchVideos();
      await useTagStore().fetchTags();
      useUiStore().pushToast({ message: 'Bulk action complete', type: 'success' });
      return response.data;
    },
    async deleteVideo(id) {
      const response = await deleteVideoApi(id);
      this.clearSelection();
      if (this.selectedVideo?.id === id) this.selectedVideo = null;
      await this.fetchVideos();
      await useTagStore().fetchTags();
      useUiStore().pushToast({ message: 'Video deleted', type: 'success' });
      return response.data;
    },
    setFilter(key, value) {
      this.filters[key] = value;
      this.fetchVideos();
    },
    setFilters(values) {
      this.filters = { ...this.filters, ...values };
      this.fetchVideos();
    },
    toggleSelect(id) {
      this.selectedIds = this.selectedIds.includes(id)
        ? this.selectedIds.filter((item) => item !== id)
        : [...this.selectedIds, id];
    },
    clearSelection() {
      this.selectedIds = [];
    },
    openVideo(video) {
      this.selectedVideo = video;
    }
  }
});
