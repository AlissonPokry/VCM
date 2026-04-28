<script setup>
import { computed, ref } from 'vue';
import { ArrowDown, ArrowUp, ArrowUpDown, Check, ChevronDown } from 'lucide-vue-next';
import { SORT_OPTIONS } from '@/constants';

const model = defineModel({ type: String, default: 'created_at' });
const order = defineModel('order', { type: String, default: 'desc' });

const open = ref(false);
const currentOption = computed(() => SORT_OPTIONS.find((option) => option.value === model.value) || SORT_OPTIONS[0]);
const orderIcon = computed(() => (order.value === 'desc' ? ArrowDown : ArrowUp));
const orderLabel = computed(() => (order.value === 'desc' ? 'DESC' : 'ASC'));

function chooseOption(value) {
  model.value = value;
  open.value = false;
}

function toggleOrder() {
  order.value = order.value === 'desc' ? 'asc' : 'desc';
}

function closeAfterFocusLeaves() {
  window.setTimeout(() => {
    open.value = false;
  }, 120);
}
</script>

<template>
  <div class="relative w-full md:w-auto" @focusout="closeAfterFocusLeaves">
    <div class="flex min-h-11 overflow-hidden rounded-md border border-border bg-elevated text-sm text-primary transition-colors focus-within:ring-2 focus-within:ring-accent hover:border-accent/70">
      <button
        type="button"
        class="flex min-w-0 flex-1 items-center gap-2 px-3 text-left transition-colors hover:bg-white/10 md:min-w-44"
        aria-haspopup="listbox"
        :aria-expanded="open"
        @click="open = !open"
      >
        <span class="flex h-7 w-7 shrink-0 items-center justify-center rounded bg-base text-secondary">
          <ArrowUpDown class="h-4 w-4" :stroke-width="1.5" />
        </span>
        <span class="min-w-0 flex-1">
          <span class="block font-mono text-[10px] uppercase tracking-[0.16em] text-secondary">Sort</span>
          <span class="block truncate font-medium text-primary">{{ currentOption.label }}</span>
        </span>
        <ChevronDown class="h-4 w-4 shrink-0 text-secondary transition-transform" :class="{ 'rotate-180': open }" :stroke-width="1.5" />
      </button>

      <span class="my-2 w-px bg-border" />

      <button
        type="button"
        class="flex min-w-20 items-center justify-center gap-1.5 px-3 font-mono text-xs text-secondary transition-colors hover:bg-white/10 hover:text-primary"
        aria-label="Toggle sort order"
        @click="toggleOrder"
      >
        <component :is="orderIcon" class="h-3.5 w-3.5" :stroke-width="1.5" />
        {{ orderLabel }}
      </button>
    </div>

    <Transition name="sort-menu">
      <div v-if="open" class="panel absolute right-0 z-30 mt-2 w-full min-w-56 rounded-md p-1 md:w-56" role="listbox">
        <button
          v-for="option in SORT_OPTIONS"
          :key="option.value"
          type="button"
          class="flex min-h-11 w-full items-center gap-3 rounded px-3 text-left text-sm transition-colors hover:bg-white/10"
          :class="model === option.value ? 'text-accent' : 'text-primary'"
          role="option"
          :aria-selected="model === option.value"
          @click="chooseOption(option.value)"
        >
          <span class="flex h-7 w-7 items-center justify-center rounded bg-base text-secondary">
            <ArrowUpDown class="h-4 w-4" :stroke-width="1.5" />
          </span>
          <span class="flex-1">{{ option.label }}</span>
          <Check v-if="model === option.value" class="h-4 w-4" :stroke-width="1.5" />
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.sort-menu-enter-active,
.sort-menu-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-fast);
}

.sort-menu-enter-from,
.sort-menu-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
