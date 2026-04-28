<script setup>
import { computed, ref } from 'vue';
import {
  Clock3,
  Film,
  HardDrive,
  Instagram,
  Music2,
  Pencil,
  Play,
  Trash2,
  Youtube
} from 'lucide-vue-next';
import { storeToRefs } from 'pinia';
import { PLATFORMS } from '@/constants';
import { useThumbnail } from '@/composables/useThumbnail';
import { useVideoStore } from '@/stores/videoStore';
import { formatDate, formatDuration, formatFileSize } from '@/utils/formatters';
import VideoStatusBadge from './VideoStatusBadge.vue';

const props = defineProps({
  video: { type: Object, required: true },
  index: { type: Number, default: 0 }
});

const emit = defineEmits(['delete', 'tag']);
const videoStore = useVideoStore();
const { selectedIds } = storeToRefs(videoStore);
const thumb = useThumbnail(props.video);
const longPressOpen = ref(false);
let touchTimer = null;

const platformIcons = { instagram: Instagram, tiktok: Music2, youtube: Youtube, all: Play };
const platform = computed(() => PLATFORMS.find((item) => item.value === props.video.platform) || PLATFORMS[0]);
const platformIcon = computed(() => platformIcons[props.video.platform] || Play);
const selected = computed(() => selectedIds.value.includes(props.video.id));
const anySelected = computed(() => selectedIds.value.length > 0);
const staggerClass = computed(() => `stagger-${Math.min(props.index, 10)}`);

function openModal() {
  videoStore.openVideo(props.video);
}

function startPress() {
  touchTimer = window.setTimeout(() => {
    longPressOpen.value = true;
  }, 400);
}

function cancelPress() {
  window.clearTimeout(touchTimer);
}
</script>

<template>
  <article
    class="video-card group rounded-md border bg-surface transition-all"
    :class="[selected ? 'border-accent shadow-glow' : 'border-border', staggerClass]"
    @touchstart.passive="startPress"
    @touchmove.passive="cancelPress"
    @touchend.passive="cancelPress"
  >
    <div class="relative aspect-video overflow-hidden rounded-t-md bg-elevated">
      <img v-if="thumb" :src="thumb" :alt="video.title" class="h-full w-full object-cover transition duration-300 ease-out group-hover:scale-[1.015] group-hover:brightness-110" />
      <div v-else class="flex h-full w-full items-center justify-center text-secondary">
        <Film class="h-12 w-12" :stroke-width="1.5" />
      </div>

      <div class="absolute left-3 top-3">
        <label class="flex h-8 w-8 translate-y-1 items-center justify-center rounded-md border border-white/20 bg-black/55 opacity-0 transition duration-300 ease-out group-hover:translate-y-0 group-hover:opacity-100" :class="{ 'opacity-100 translate-y-0': anySelected }">
          <input
            class="h-4 w-4 accent-[var(--color-accent)]"
            type="checkbox"
            :checked="selected"
            @click.stop="videoStore.toggleSelect(video.id)"
          />
        </label>
      </div>

      <div class="absolute right-3 top-3">
        <VideoStatusBadge :status="video.status" />
      </div>

      <button class="absolute inset-0 flex items-center justify-center bg-gradient-to-t from-black/70 via-black/20 to-transparent opacity-0 transition duration-300 ease-out group-hover:opacity-100" type="button" @click="openModal">
        <Play class="h-10 w-10 text-white" :stroke-width="1.5" />
      </button>

      <div class="absolute bottom-3 right-3 flex translate-y-2 gap-2 opacity-0 transition duration-300 ease-out group-hover:translate-y-0 group-hover:opacity-100">
        <button class="icon-button h-9 w-9 bg-black/60" type="button" aria-label="Edit video" @click.stop="openModal">
          <Pencil class="h-4 w-4" :stroke-width="1.5" />
        </button>
        <button class="icon-button h-9 w-9 bg-black/60 hover:text-warm" type="button" aria-label="Delete video" @click.stop="emit('delete', video)">
          <Trash2 class="h-4 w-4" :stroke-width="1.5" />
        </button>
      </div>
    </div>

    <button type="button" class="block w-full p-4 text-left" @click="openModal">
      <h3 class="truncate font-display text-lg font-semibold">{{ video.title }}</h3>
      <p class="mt-1 line-clamp-2 min-h-10 text-sm text-secondary">{{ video.description || 'No description yet.' }}</p>
      <div class="mt-4 grid grid-cols-2 gap-2 font-mono text-[11px] text-secondary">
        <span class="flex min-w-0 items-center gap-1 truncate">
          <component :is="platformIcon" class="h-3.5 w-3.5 shrink-0" :stroke-width="1.5" />
          {{ platform.label }}
        </span>
        <span class="truncate">{{ formatDate(video.status === 'posted' ? video.posted_at : video.scheduled_at) }}</span>
        <span class="flex items-center gap-1">
          <HardDrive class="h-3.5 w-3.5" :stroke-width="1.5" />
          {{ formatFileSize(video.file_size) }}
        </span>
        <span class="flex items-center gap-1">
          <Clock3 class="h-3.5 w-3.5" :stroke-width="1.5" />
          {{ formatDuration(video.duration) }}
        </span>
      </div>
    </button>

    <div class="flex flex-wrap gap-2 px-4 pb-4">
      <button
        v-for="tag in video.tags"
        :key="tag"
        type="button"
        class="rounded-full bg-accent/10 px-2 py-1 text-xs text-accent hover:bg-accent hover:text-white"
        @click.stop="emit('tag', tag)"
      >
        {{ tag }}
      </button>
    </div>

    <div v-if="longPressOpen" class="flex gap-2 border-t border-border p-3 md:hidden">
      <button type="button" class="btn-secondary flex-1" @click="openModal">Edit</button>
      <button type="button" class="btn-danger flex-1" @click="emit('delete', video)">Delete</button>
    </div>
  </article>
</template>
