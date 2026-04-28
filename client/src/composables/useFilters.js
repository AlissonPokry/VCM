import { computed } from 'vue';
import { useVideoStore } from '@/stores/videoStore';

export function useFilters() {
  const videoStore = useVideoStore();
  const filters = computed(() => videoStore.filters);
  const activeCount = computed(() => {
    const f = filters.value;
    return [
      f.platform !== 'all',
      f.search,
      f.tags.length,
      f.dateFrom,
      f.dateTo,
      f.sort !== 'created_at',
      f.order !== 'desc'
    ].filter(Boolean).length;
  });
  const setFilter = (key, value) => videoStore.setFilter(key, value);
  const clearFilters = () => videoStore.setFilters({
    platform: 'all',
    search: '',
    tags: [],
    dateFrom: null,
    dateTo: null,
    sort: 'created_at',
    order: 'desc'
  });

  return { filters, activeCount, setFilter, clearFilters };
}
