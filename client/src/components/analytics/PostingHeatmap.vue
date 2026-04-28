<script setup>
import { computed } from 'vue';
import { lastNDays } from '@/utils/formatters';

const props = defineProps({
  data: { type: Array, default: () => [] }
});

const days = computed(() => {
  const counts = new Map(props.data.map((row) => [row.date, Number(row.count || 0)]));
  return lastNDays(90).map((date) => {
    const count = counts.get(date) || 0;
    return {
      date,
      count,
      heat: count === 0 ? 0 : count === 1 ? 1 : count === 2 ? 2 : 3
    };
  });
});
</script>

<template>
  <section class="panel rounded-md p-4">
    <div class="mb-4 flex items-center justify-between">
      <h2 class="font-display text-xl font-semibold">Posting heatmap</h2>
      <p class="font-mono text-xs text-secondary">90 days</p>
    </div>
    <div class="overflow-x-auto pb-2">
      <div class="grid w-max grid-flow-col grid-rows-7 gap-1">
        <span
          v-for="day in days"
          :key="day.date"
          class="h-4 w-4 rounded-[3px]"
          :class="`heat-${day.heat}`"
          :title="`${day.date}: ${day.count} post${day.count === 1 ? '' : 's'}`"
        />
      </div>
    </div>
  </section>
</template>
