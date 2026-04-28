<script setup>
import { ref } from 'vue';
import { Film, Upload, X } from 'lucide-vue-next';
import TagInput from '@/components/shared/TagInput.vue';
import { PLATFORMS } from '@/constants';
import { useVideoUpload } from '@/composables/useVideoUpload';
import { formatFileSize } from '@/utils/formatters';
import UploadProgress from './UploadProgress.vue';

const emit = defineEmits(['upload-complete']);
const input = ref(null);
const dragging = ref(false);
const { files, addFiles, removeFile, uploadAll, progress, uploading } = useVideoUpload();

function handleFiles(fileList) {
  addFiles(fileList);
}

async function submit() {
  await uploadAll();
  emit('upload-complete');
}
</script>

<template>
  <div class="space-y-4">
    <button
      type="button"
      class="flex min-h-[260px] w-full flex-col items-center justify-center rounded-md border border-dashed p-8 text-center transition-colors"
      :class="dragging ? 'border-accent bg-accent/5' : 'border-border bg-surface/60 hover:border-accent/70'"
      @click="input?.click()"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop.prevent="dragging = false; handleFiles($event.dataTransfer.files)"
    >
      <Film class="mb-4 h-14 w-14 text-secondary transition-transform" :class="{ 'scale-110 text-accent': dragging }" :stroke-width="1.5" />
      <span class="font-display text-2xl font-semibold">Drop vertical videos</span>
      <span class="mt-2 text-sm text-secondary">Or click to browse. Multiple files supported.</span>
      <input ref="input" type="file" accept="video/*" multiple class="sr-only" @change="handleFiles($event.target.files)" />
    </button>

    <div v-if="files.length" class="space-y-3">
      <article v-for="item in files" :key="item.id" class="panel rounded-md p-4">
        <div class="mb-4 flex items-start gap-3">
          <div class="flex h-11 w-11 shrink-0 items-center justify-center rounded-md bg-accent/10 text-accent">
            <Upload class="h-5 w-5" :stroke-width="1.5" />
          </div>
          <div class="min-w-0 flex-1">
            <p class="truncate font-display text-lg font-semibold">{{ item.file.name }}</p>
            <p class="font-mono text-xs text-secondary">{{ formatFileSize(item.file.size) }}</p>
          </div>
          <button class="icon-button" type="button" aria-label="Remove file" @click="removeFile(item.id)">
            <X class="h-4 w-4" :stroke-width="1.5" />
          </button>
        </div>

        <div class="grid gap-3 lg:grid-cols-2">
          <label>
            <span class="mb-1 block text-xs uppercase text-secondary">Title</span>
            <input v-model="item.title" class="control w-full" />
          </label>
          <label>
            <span class="mb-1 block text-xs uppercase text-secondary">Platform</span>
            <select v-model="item.platform" class="control w-full">
              <option v-for="platform in PLATFORMS.filter((entry) => entry.value !== 'all')" :key="platform.value" :value="platform.value">
                {{ platform.label }}
              </option>
            </select>
          </label>
          <label>
            <span class="mb-1 block text-xs uppercase text-secondary">Schedule</span>
            <input v-model="item.scheduled_at" type="datetime-local" class="control w-full" />
          </label>
          <div>
            <span class="mb-1 block text-xs uppercase text-secondary">Tags</span>
            <TagInput v-model="item.tags" />
          </div>
          <label class="lg:col-span-2">
            <span class="mb-1 block text-xs uppercase text-secondary">Description</span>
            <textarea v-model="item.description" class="control min-h-24 w-full py-3" />
          </label>
        </div>

        <UploadProgress class="mt-4" :progress="progress[item.id] || 0" :complete="progress[item.id] === 100" />
      </article>
      <button type="button" class="btn-primary w-full" :disabled="uploading" @click="submit">
        Upload {{ files.length }} file{{ files.length === 1 ? '' : 's' }}
      </button>
    </div>
  </div>
</template>
