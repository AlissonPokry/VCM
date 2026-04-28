<script setup>
import { ref } from 'vue';
import { CalendarArrowUp, Check, Trash2, X } from 'lucide-vue-next';
import { storeToRefs } from 'pinia';
import ConfirmDialog from '@/components/shared/ConfirmDialog.vue';
import { useVideoStore } from '@/stores/videoStore';
import { toIsoFromInput } from '@/utils/formatters';

const videoStore = useVideoStore();
const { selectedIds } = storeToRefs(videoStore);
const rescheduleAt = ref('');
const confirmDelete = ref(false);
</script>

<template>
  <Transition name="bulk">
    <div v-if="selectedIds.length" class="fixed inset-x-0 bottom-0 z-50 border-t border-border bg-elevated/95 p-3 shadow-2xl backdrop-blur">
      <div class="mx-auto flex max-w-5xl flex-wrap items-center gap-3">
        <button class="icon-button" type="button" aria-label="Deselect all" @click="videoStore.clearSelection()">
          <X class="h-4 w-4" :stroke-width="1.5" />
        </button>
        <p class="mr-auto font-display text-lg font-semibold">{{ selectedIds.length }} selected</p>
        <input v-model="rescheduleAt" type="datetime-local" class="control" />
        <button class="btn-secondary flex-1 sm:flex-none" type="button" @click="videoStore.bulkAction(selectedIds, 'reschedule', { scheduled_at: toIsoFromInput(rescheduleAt) })">
          <CalendarArrowUp class="h-4 w-4" :stroke-width="1.5" />
          Reschedule
        </button>
        <button class="btn-secondary flex-1 sm:flex-none" type="button" @click="videoStore.bulkAction(selectedIds, 'draft')">
          <Check class="h-4 w-4" :stroke-width="1.5" />
          Draft
        </button>
        <button class="btn-danger flex-1 sm:flex-none" type="button" @click="confirmDelete = true">
          <Trash2 class="h-4 w-4" :stroke-width="1.5" />
          Delete
        </button>
      </div>
    </div>
  </Transition>

  <ConfirmDialog
    v-if="confirmDelete"
    title="Delete selected videos"
    :message="`Delete ${selectedIds.length} videos, including files and thumbnails?`"
    confirm-label="Delete"
    destructive
    @cancel="confirmDelete = false"
    @confirm="videoStore.bulkAction(selectedIds, 'delete'); confirmDelete = false"
  />
</template>

<style scoped>
.bulk-enter-active {
  transition: transform var(--transition-spring);
}

.bulk-leave-active {
  transition: transform var(--transition-base);
}

.bulk-enter-from,
.bulk-leave-to {
  transform: translateY(100%);
}
</style>
