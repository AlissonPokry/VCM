import { defineStore } from 'pinia';
import { getActivity } from '@/api/activity';

export const useActivityStore = defineStore('activity', {
  state: () => ({
    entries: [],
    total: 0,
    loading: false,
    limit: 50,
    offset: 0
  }),
  actions: {
    async fetchActivity(opts = {}) {
      this.loading = true;
      try {
        const response = await getActivity({ limit: this.limit, offset: 0, ...opts });
        this.entries = response.data;
        this.total = response.meta?.total || response.data.length;
        this.offset = response.data.length;
      } finally {
        this.loading = false;
      }
    },
    async loadMore(opts = {}) {
      const response = await getActivity({ limit: this.limit, offset: this.offset, ...opts });
      this.entries = [...this.entries, ...response.data];
      this.total = response.meta?.total || this.total;
      this.offset += response.data.length;
    }
  }
});
