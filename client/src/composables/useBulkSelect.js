import { computed } from 'vue';
import { useVideoStore } from '@/stores/videoStore';

export function useBulkSelect() {
  const videoStore = useVideoStore();
  const selectedIds = computed(() => videoStore.selectedIds);
  const isSelected = (id) => selectedIds.value.includes(id);
  const toggle = (id) => videoStore.toggleSelect(id);
  const clearAll = () => videoStore.clearSelection();
  const selectAll = (ids) => {
    videoStore.selectedIds = [...new Set(ids)];
  };

  return { selectedIds, toggle, clearAll, isSelected, selectAll };
}
