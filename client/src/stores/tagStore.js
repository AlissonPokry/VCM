import { defineStore } from 'pinia';
import { getTags } from '@/api/tags';

export const useTagStore = defineStore('tags', {
  state: () => ({
    tags: []
  }),
  actions: {
    async fetchTags() {
      const response = await getTags();
      this.tags = response.data;
    }
  }
});
