<script setup>
import { computed, ref } from 'vue';
import { Tag, X } from 'lucide-vue-next';
import { useTagStore } from '@/stores/tagStore';

const model = defineModel({ type: Array, default: () => [] });
const tagStore = useTagStore();
const query = ref('');
const focused = ref(false);

const matches = computed(() => {
  const q = query.value.trim().toLowerCase();
  return tagStore.tags
    .filter((tag) => !model.value.some((name) => name.toLowerCase() === tag.name.toLowerCase()))
    .filter((tag) => !q || tag.name.toLowerCase().includes(q))
    .slice(0, 8);
});

const canCreate = computed(() => {
  const q = query.value.trim();
  return q && !tagStore.tags.some((tag) => tag.name.toLowerCase() === q.toLowerCase()) && !model.value.some((tag) => tag.toLowerCase() === q.toLowerCase());
});

function addTag(name) {
  const clean = name.trim().replace(/,$/, '');
  if (!clean || model.value.some((tag) => tag.toLowerCase() === clean.toLowerCase())) return;
  model.value = [...model.value, clean];
  query.value = '';
}

function removeTag(name) {
  model.value = model.value.filter((tag) => tag !== name);
}

function handleKey(event) {
  if (event.key === 'Enter' || event.key === ',') {
    event.preventDefault();
    addTag(query.value);
  }
  if (event.key === 'Backspace' && !query.value && model.value.length) {
    removeTag(model.value[model.value.length - 1]);
  }
}

function handleBlur() {
  window.setTimeout(() => {
    focused.value = false;
  }, 120);
}
</script>

<template>
  <div class="relative">
    <div class="flex min-h-11 flex-wrap items-center gap-2 rounded-md border border-border bg-elevated px-3 py-2 focus-within:ring-2 focus-within:ring-accent">
      <Tag class="h-4 w-4 text-secondary" :stroke-width="1.5" />
      <span v-for="tag in model" :key="tag" class="inline-flex items-center gap-1 rounded bg-accent/10 px-2 py-1 text-xs text-accent">
        {{ tag }}
        <button type="button" :aria-label="`Remove ${tag}`" @click="removeTag(tag)">
          <X class="h-3 w-3" :stroke-width="1.5" />
        </button>
      </span>
      <input
        v-model="query"
        class="min-w-28 flex-1 bg-transparent text-sm text-primary outline-none placeholder:text-secondary"
        placeholder="Add tags"
        @focus="focused = true"
        @blur="handleBlur"
        @keydown="handleKey"
      />
    </div>
    <div v-if="focused && (matches.length || canCreate)" class="panel absolute z-20 mt-2 max-h-64 w-full overflow-y-auto rounded-md p-1">
      <button
        v-for="tag in matches"
        :key="tag.id"
        type="button"
        class="flex min-h-11 w-full items-center justify-between rounded px-3 text-left text-sm text-primary hover:bg-white/10"
        @mousedown.prevent="addTag(tag.name)"
      >
        <span>{{ tag.name }}</span>
        <span class="font-mono text-xs text-secondary">{{ tag.count }}</span>
      </button>
      <button
        v-if="canCreate"
        type="button"
        class="min-h-11 w-full rounded px-3 text-left text-sm text-accent hover:bg-accent/10"
        @mousedown.prevent="addTag(query)"
      >
        + Create "{{ query.trim() }}"
      </button>
    </div>
  </div>
</template>
