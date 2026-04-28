<script setup>
import { computed } from 'vue';

const props = defineProps({
  label: { type: String, required: true },
  value: { type: [String, Number], default: null },
  trend: { type: String, default: null },
  unit: { type: String, default: '' },
  icon: { type: [Object, Function], required: true }
});

const trendClass = computed(() => (String(props.trend || '').startsWith('-') ? 'text-warm' : 'text-green'));
const trendArrow = computed(() => (String(props.trend || '').startsWith('-') ? '↓' : '↑'));
</script>

<template>
  <article class="panel rounded-md p-4">
    <div v-if="value === null" class="space-y-3">
      <div class="h-6 w-10 rounded skeleton" />
      <div class="h-4 rounded skeleton" />
    </div>
    <template v-else>
      <div class="mb-4 flex items-center justify-between text-secondary">
        <component :is="icon" class="h-5 w-5" :stroke-width="1.5" />
        <span v-if="trend" class="font-mono text-xs" :class="trendClass">{{ trendArrow }} {{ trend }}</span>
      </div>
      <p class="font-display text-3xl font-semibold">{{ value }}<span v-if="unit" class="text-lg text-secondary">{{ unit }}</span></p>
      <p class="mt-1 text-sm text-secondary">{{ label }}</p>
    </template>
  </article>
</template>
