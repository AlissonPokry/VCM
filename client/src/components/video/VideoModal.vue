<script setup>
import { computed, reactive, ref, watch } from 'vue';
import { Check, Trash2, X } from 'lucide-vue-next';
import ActivityFeed from '@/components/activity/ActivityFeed.vue';
import ConfirmDialog from '@/components/shared/ConfirmDialog.vue';
import TagInput from '@/components/shared/TagInput.vue';
import { PLATFORMS } from '@/constants';
import { videoFileUrl } from '@/api/videos';
import { useVideoStore } from '@/stores/videoStore';
import { formatDateInput, toIsoFromInput } from '@/utils/formatters';

const props = defineProps({
  video: { type: Object, required: true }
});

const emit = defineEmits(['close']);
const videoStore = useVideoStore();
const confirmDelete = ref(false);
const touchStartY = ref(0);
const form = reactive({
  title: '',
  description: '',
  platform: 'instagram',
  scheduled_at: '',
  tags: []
});

const statusAction = computed(() => (props.video.status === 'posted' ? 'draft' : 'posted'));
const statusLabel = computed(() => (props.video.status === 'posted' ? 'Move to Draft' : 'Mark as Posted'));

watch(
  () => props.video,
  (video) => {
    form.title = video.title || '';
    form.description = video.description || '';
    form.platform = video.platform || 'instagram';
    form.scheduled_at = formatDateInput(video.scheduled_at);
    form.tags = [...(video.tags || [])];
  },
  { immediate: true }
);

function save() {
  videoStore.updateVideo(props.video.id, {
    title: form.title,
    description: form.description,
    platform: form.platform,
    scheduled_at: toIsoFromInput(form.scheduled_at),
    tags: form.tags
  });
}

function setStatus() {
  videoStore.updateStatus(props.video.id, statusAction.value);
}

function deleteCurrent() {
  videoStore.deleteVideo(props.video.id);
  confirmDelete.value = false;
  emit('close');
}

function onTouchStart(event) {
  touchStartY.value = event.touches[0].clientY;
}

function onTouchEnd(event) {
  if (event.changedTouches[0].clientY - touchStartY.value > 90) {
    emit('close');
  }
}
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div class="fixed inset-0 z-[70]">
        <button class="absolute inset-0 bg-black/60" aria-label="Close video panel" @click="emit('close')" />
        <section
          class="panel absolute bottom-0 right-0 flex h-[90vh] w-full flex-col overflow-hidden rounded-t-md lg:top-0 lg:h-full lg:w-[480px] lg:rounded-none"
          @touchstart.passive="onTouchStart"
          @touchend.passive="onTouchEnd"
        >
          <div class="flex items-center justify-between border-b border-border p-4">
            <div class="min-w-0">
              <p class="font-mono text-xs uppercase tracking-[0.16em] text-secondary">Video detail</p>
              <h2 class="truncate font-display text-xl font-semibold">{{ video.title }}</h2>
            </div>
            <button class="icon-button" type="button" aria-label="Close video panel" @click="emit('close')">
              <X class="h-4 w-4" :stroke-width="1.5" />
            </button>
          </div>

          <div class="flex-1 overflow-y-auto p-4">
            <video class="mb-4 aspect-video w-full rounded-md bg-black" controls :src="videoFileUrl(video.id)" />

            <div class="space-y-4">
              <label class="block">
                <span class="mb-1 block text-xs uppercase text-secondary">Title</span>
                <input v-model="form.title" class="control w-full" />
              </label>
              <label class="block">
                <span class="mb-1 block text-xs uppercase text-secondary">Description</span>
                <textarea v-model="form.description" class="control min-h-28 w-full py-3" />
              </label>
              <label class="block">
                <span class="mb-1 block text-xs uppercase text-secondary">Platform</span>
                <select v-model="form.platform" class="control w-full">
                  <option v-for="platform in PLATFORMS.filter((item) => item.value !== 'all')" :key="platform.value" :value="platform.value">
                    {{ platform.label }}
                  </option>
                </select>
              </label>
              <label class="block">
                <span class="mb-1 block text-xs uppercase text-secondary">Schedule</span>
                <input v-model="form.scheduled_at" type="datetime-local" class="control w-full" />
              </label>
              <div>
                <span class="mb-1 block text-xs uppercase text-secondary">Tags</span>
                <TagInput v-model="form.tags" />
              </div>
            </div>

            <div class="mt-6">
              <h3 class="mb-3 font-display text-lg font-semibold">Activity</h3>
              <div class="max-h-80">
                <ActivityFeed :video-id="video.id" compact />
              </div>
            </div>
          </div>

          <div class="grid grid-cols-3 gap-2 border-t border-border p-4">
            <button class="btn-primary" type="button" @click="save">
              <Check class="h-4 w-4" :stroke-width="1.5" />
              Save
            </button>
            <button class="btn-secondary" type="button" @click="setStatus">{{ statusLabel }}</button>
            <button class="btn-danger" type="button" @click="confirmDelete = true">
              <Trash2 class="h-4 w-4" :stroke-width="1.5" />
              Delete
            </button>
          </div>
        </section>
      </div>
    </Transition>

    <ConfirmDialog
      v-if="confirmDelete"
      title="Delete video"
      :message="`Delete ${video.title} permanently?`"
      confirm-label="Delete"
      destructive
      @cancel="confirmDelete = false"
      @confirm="deleteCurrent"
    />
  </Teleport>
</template>

<style scoped>
.modal-enter-active,
.modal-leave-active {
  transition: opacity var(--transition-slow);
}

.modal-enter-active section,
.modal-leave-active section {
  transition: transform var(--transition-slow);
  will-change: transform;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from section,
.modal-leave-to section {
  transform: translate3d(0, 100%, 0);
}

@media (min-width: 1024px) {
  .modal-enter-from section,
  .modal-leave-to section {
    transform: translate3d(100%, 0, 0);
  }
}
</style>
