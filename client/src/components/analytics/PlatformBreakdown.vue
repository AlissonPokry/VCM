<script setup>
import { computed } from 'vue';
import { PLATFORMS } from '@/constants';

const props = defineProps({
  breakdown: { type: Object, default: () => ({}) }
});

const total = computed(() => Object.values(props.breakdown || {}).reduce((sum, count) => sum + Number(count || 0), 0));
const segments = computed(() => PLATFORMS.filter((item) => item.value !== 'all').map((platform) => {
  const count = Number(props.breakdown?.[platform.value] || 0);
  const percent = total.value ? Math.round((count / total.value) * 100) : 0;
  const bucket = Math.max(count ? 5 : 0, Math.round(percent / 5) * 5);
  return { ...platform, count, percent, bucket };
}));
</script>

<template>
  <section class="panel rounded-md p-4">
    <h2 class="mb-4 font-display text-xl font-semibold">Platform breakdown</h2>
    <div class="flex h-12 overflow-hidden rounded-md bg-base">
      <div
        v-for="segment in segments"
        :key="segment.value"
        class="flex min-w-0 items-center justify-center font-mono text-xs text-black"
        :class="[`platform-${segment.value}`, `pct-${segment.bucket}`]"
        :title="`${segment.label}: ${segment.count}`"
      >
        <span v-if="segment.percent >= 12">{{ segment.percent }}%</span>
      </div>
    </div>
    <div class="mt-4 grid gap-2">
      <div v-for="segment in segments" :key="segment.value" class="flex items-center gap-2 text-sm text-secondary">
        <span class="h-2.5 w-2.5 rounded-full" :class="`platform-${segment.value}`" />
        <span class="text-primary">{{ segment.label }}</span>
        <span class="ml-auto font-mono">{{ segment.count }} / {{ segment.percent }}%</span>
      </div>
    </div>
  </section>
</template>
