<script setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { Menu } from 'lucide-vue-next';

defineEmits(['toggle-sidebar']);

const route = useRoute();
const title = computed(() => route.name || 'Scheduled');
</script>

<template>
  <header class="sticky top-0 z-30 border-b border-border bg-base/80 px-4 py-3 backdrop-blur sm:px-6 lg:px-8">
    <div class="mx-auto flex max-w-[1520px] items-center gap-3">
      <button class="icon-button md:hidden" type="button" aria-label="Open navigation" @click="$emit('toggle-sidebar')">
        <Menu class="h-5 w-5" :stroke-width="1.5" />
      </button>
      <div>
        <p class="font-mono text-xs uppercase tracking-[0.18em] text-secondary">Vertical Content Manager</p>
        <h1 class="font-display text-2xl font-semibold text-primary">{{ title }}</h1>
      </div>
      <div class="ml-auto">
        <slot name="actions" />
      </div>
    </div>
  </header>
</template>
