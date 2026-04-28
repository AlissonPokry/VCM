<script setup>
import { onMounted, onUnmounted } from 'vue';
import {
  AlertTriangle,
  CheckCircle2,
  FilePen,
  Send,
  Trash2,
  Upload,
  Zap
} from 'lucide-vue-next';
import { storeToRefs } from 'pinia';
import { ACTIVITY_ACTIONS } from '@/constants';
import { useActivityStore } from '@/stores/activityStore';
import { useVideoStore } from '@/stores/videoStore';
import { relativeTime } from '@/utils/formatters';

const props = defineProps({
  videoId: { type: [Number, String], default: null },
  compact: { type: Boolean, default: false }
});

const icons = { AlertTriangle, CheckCircle2, FilePen, Send, Trash2, Upload, Zap };
const activityStore = useActivityStore();
const videoStore = useVideoStore();
const { entries, total, loading } = storeToRefs(activityStore);
let timer = null;

function iconFor(action) {
  return icons[ACTIVITY_ACTIONS[action]?.icon] || Zap;
}

function colorFor(action) {
  return ACTIVITY_ACTIONS[action]?.color || 'text-secondary';
}

async function refresh() {
  await activityStore.fetchActivity(props.videoId ? { video_id: props.videoId } : {});
}

async function openVideo(entry) {
  if (!entry.video_id) return;
  await videoStore.fetchVideo(entry.video_id);
}

onMounted(() => {
  refresh().catch(console.error);
  timer = window.setInterval(() => refresh().catch(console.error), 30000);
});

onUnmounted(() => {
  window.clearInterval(timer);
});
</script>

<template>
  <div class="flex h-full flex-col">
    <div v-if="loading" class="space-y-3">
      <div v-for="n in 5" :key="n" class="h-14 rounded-md skeleton" />
    </div>
    <ul v-else class="flex-1 space-y-3 overflow-y-auto pr-1">
      <li v-for="entry in entries" :key="entry.id" class="flex gap-3 rounded-md border border-border bg-elevated/50 p-3">
        <div class="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-white/5" :class="colorFor(entry.action)">
          <component :is="iconFor(entry.action)" class="h-4 w-4" :stroke-width="1.5" />
        </div>
        <div class="min-w-0 flex-1">
          <button
            v-if="entry.video_id"
            type="button"
            class="mb-1 max-w-full truncate text-left text-sm text-primary hover:text-accent"
            @click="openVideo(entry)"
          >
            {{ entry.video_title || 'Deleted video' }}
          </button>
          <p class="text-sm text-secondary">{{ entry.detail }}</p>
          <div class="mt-2 flex items-center gap-2 font-mono text-[11px] text-muted">
            <span>{{ relativeTime(entry.created_at) }}</span>
            <span v-if="entry.source === 'n8n'" class="rounded bg-accent/10 px-1.5 py-0.5 text-accent">n8n</span>
          </div>
        </div>
      </li>
    </ul>
    <button
      v-if="!compact && entries.length < total"
      type="button"
      class="btn-secondary mt-4 w-full"
      @click="activityStore.loadMore(props.videoId ? { video_id: props.videoId } : {})"
    >
      Load more
    </button>
  </div>
</template>
