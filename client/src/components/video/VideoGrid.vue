<script setup>
import { Inbox } from 'lucide-vue-next';
import EmptyState from '@/components/shared/EmptyState.vue';
import VideoCard from './VideoCard.vue';

defineProps({
  videos: { type: Array, required: true },
  loading: { type: Boolean, default: false },
  emptyTitle: { type: String, default: 'No videos found' },
  emptyDescription: { type: String, default: 'Upload or adjust filters to build this queue.' }
});

defineEmits(['delete', 'tag']);
</script>

<template>
  <div v-if="loading" class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
    <div v-for="n in 8" :key="n" class="overflow-hidden rounded-md border border-border bg-surface">
      <div class="aspect-video skeleton" />
      <div class="space-y-3 p-4">
        <div class="h-5 rounded skeleton" />
        <div class="h-4 rounded skeleton" />
        <div class="h-4 w-2/3 rounded skeleton" />
      </div>
    </div>
  </div>
  <EmptyState v-else-if="videos.length === 0" :icon="Inbox" :title="emptyTitle" :description="emptyDescription" />
  <div v-else class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
    <VideoCard
      v-for="(video, index) in videos"
      :key="video.id"
      :video="video"
      :index="index"
      @delete="$emit('delete', $event)"
      @tag="$emit('tag', $event)"
    />
  </div>
</template>
