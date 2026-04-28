<script setup>
import { onMounted, ref } from 'vue';
import ConfirmDialog from '@/components/shared/ConfirmDialog.vue';
import FilterBar from '@/components/shared/FilterBar.vue';
import BulkActionBar from '@/components/video/BulkActionBar.vue';
import VideoGrid from '@/components/video/VideoGrid.vue';
import { useVideoStore } from '@/stores/videoStore';

const videoStore = useVideoStore();
const pendingDelete = ref(null);

onMounted(() => {
  videoStore.setFilters({ status: 'posted' });
});

function applyTag(tag) {
  videoStore.setFilter('tags', [tag]);
}
</script>

<template>
  <section>
    <FilterBar />
    <VideoGrid
      :videos="videoStore.videos"
      :loading="videoStore.loading"
      empty-title="No posted videos"
      empty-description="Posted videos appear here after manual status changes or n8n success webhooks."
      @delete="pendingDelete = $event"
      @tag="applyTag"
    />
    <BulkActionBar />
    <ConfirmDialog
      v-if="pendingDelete"
      title="Delete video"
      :message="`Delete ${pendingDelete.title} permanently?`"
      confirm-label="Delete"
      destructive
      @cancel="pendingDelete = null"
      @confirm="videoStore.deleteVideo(pendingDelete.id); pendingDelete = null"
    />
  </section>
</template>
