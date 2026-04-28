<script setup>
import { computed, ref, watch } from 'vue';
import { Search, SlidersHorizontal, X } from 'lucide-vue-next';
import { PLATFORMS } from '@/constants';
import { useFilters } from '@/composables/useFilters';
import { useTagStore } from '@/stores/tagStore';
import SortDropdown from './SortDropdown.vue';

const { filters, activeCount, setFilter, clearFilters } = useFilters();
const tagStore = useTagStore();
const mobileOpen = ref(false);
const search = ref(filters.value.search);
let debounceTimer = null;

watch(search, (value) => {
  window.clearTimeout(debounceTimer);
  debounceTimer = window.setTimeout(() => setFilter('search', value), 300);
});

const tagOptions = computed(() => tagStore.tags);
const activePlatformIndex = computed(() => Math.max(0, PLATFORMS.findIndex((platform) => platform.value === filters.value.platform)));
const activePlatformClass = computed(() => `platform-pill-${activePlatformIndex.value}`);

function toggleTag(name) {
  const tags = filters.value.tags.includes(name)
    ? filters.value.tags.filter((tag) => tag !== name)
    : [...filters.value.tags, name];
  setFilter('tags', tags);
}
</script>

<template>
  <div class="mb-5">
    <button class="btn-secondary w-full justify-between md:hidden" type="button" @click="mobileOpen = true">
      <span class="flex items-center gap-2">
        <SlidersHorizontal class="h-4 w-4" :stroke-width="1.5" />
        Filters
      </span>
      <span v-if="activeCount" class="rounded bg-accent px-2 py-0.5 text-xs text-white">{{ activeCount }}</span>
    </button>

    <div class="hidden items-center gap-3 md:flex">
      <label class="control flex flex-1 items-center gap-2">
        <Search class="h-4 w-4 text-secondary" :stroke-width="1.5" />
        <input v-model="search" class="w-full bg-transparent outline-none" placeholder="Search videos" />
      </label>
      <div class="relative grid min-h-11 grid-cols-4 overflow-hidden rounded-md border border-border bg-elevated p-1">
        <span
          class="pointer-events-none absolute bottom-1 top-1 w-1/4 rounded bg-accent shadow-glow"
          :class="activePlatformClass"
        />
        <button
          v-for="platform in PLATFORMS"
          :key="platform.value"
          type="button"
          class="relative z-10 min-h-9 min-w-20 rounded px-3 text-sm font-medium transition duration-300 ease-out"
          :class="filters.platform === platform.value ? 'text-white' : 'text-secondary hover:text-primary'"
          @click="setFilter('platform', platform.value)"
        >
          {{ platform.label }}
        </button>
      </div>
      <input :value="filters.dateFrom" type="date" class="control" @input="setFilter('dateFrom', $event.target.value || null)" />
      <input :value="filters.dateTo" type="date" class="control" @input="setFilter('dateTo', $event.target.value || null)" />
      <SortDropdown
        :model-value="filters.sort"
        :order="filters.order"
        @update:model-value="setFilter('sort', $event)"
        @update:order="setFilter('order', $event)"
      />
    </div>

    <TransitionGroup name="chip" tag="div" class="mt-3 flex flex-wrap gap-2">
      <button
        v-for="tag in tagOptions"
        :key="tag.id"
        type="button"
        class="rounded-full border px-3 py-1 text-xs transition-colors"
        :class="filters.tags.includes(tag.name) ? 'border-accent bg-accent/15 text-accent' : 'border-border text-secondary hover:text-primary'"
        @click="toggleTag(tag.name)"
      >
        {{ tag.name }} {{ tag.count }}
      </button>
    </TransitionGroup>

    <Teleport to="body">
      <Transition name="sheet">
        <div v-if="mobileOpen" class="fixed inset-0 z-[75] md:hidden">
          <button class="absolute inset-0 bg-black/70" aria-label="Close filters" @click="mobileOpen = false" />
          <section class="panel absolute bottom-0 left-0 right-0 max-h-[88vh] overflow-y-auto rounded-t-md p-4">
            <div class="mb-4 flex items-center justify-between">
              <h2 class="font-display text-xl font-semibold">Filters</h2>
              <button class="icon-button" type="button" aria-label="Close filters" @click="mobileOpen = false">
                <X class="h-4 w-4" :stroke-width="1.5" />
              </button>
            </div>
            <div class="space-y-3">
              <label class="control flex items-center gap-2">
                <Search class="h-4 w-4 text-secondary" :stroke-width="1.5" />
                <input v-model="search" class="w-full bg-transparent outline-none" placeholder="Search videos" />
              </label>
              <div class="grid grid-cols-2 gap-2">
                <button
                  v-for="platform in PLATFORMS"
                  :key="platform.value"
                  type="button"
                  class="btn-secondary"
                  :class="filters.platform === platform.value ? 'bg-accent text-white' : ''"
                  @click="setFilter('platform', platform.value)"
                >
                  {{ platform.label }}
                </button>
              </div>
              <input :value="filters.dateFrom" type="date" class="control w-full" @input="setFilter('dateFrom', $event.target.value || null)" />
              <input :value="filters.dateTo" type="date" class="control w-full" @input="setFilter('dateTo', $event.target.value || null)" />
              <SortDropdown
                :model-value="filters.sort"
                :order="filters.order"
                @update:model-value="setFilter('sort', $event)"
                @update:order="setFilter('order', $event)"
              />
              <button type="button" class="btn-secondary w-full" @click="clearFilters">Clear filters</button>
            </div>
          </section>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.chip-enter-active,
.chip-leave-active,
.sheet-enter-active,
.sheet-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}

.chip-enter-from,
.chip-leave-to {
  opacity: 0;
  transform: scale(0.8);
}

.sheet-enter-from,
.sheet-leave-to {
  opacity: 0;
  transform: translateY(12px);
}

.platform-pill-0 {
  transform: translate3d(0, 0, 0);
  transition: transform var(--transition-spring), box-shadow var(--transition-base);
}

.platform-pill-1 {
  transform: translate3d(100%, 0, 0);
  transition: transform var(--transition-spring), box-shadow var(--transition-base);
}

.platform-pill-2 {
  transform: translate3d(200%, 0, 0);
  transition: transform var(--transition-spring), box-shadow var(--transition-base);
}

.platform-pill-3 {
  transform: translate3d(300%, 0, 0);
  transition: transform var(--transition-spring), box-shadow var(--transition-base);
}
</style>
