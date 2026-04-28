<script setup>
import { computed, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  BarChart3,
  CalendarClock,
  CheckCircle2,
  Database,
  FileVideo,
  Flame,
  Hash,
  Layers
} from 'lucide-vue-next';
import { getHeatmap, getSummary } from '@/api/analytics';
import PostingHeatmap from '@/components/analytics/PostingHeatmap.vue';
import PlatformBreakdown from '@/components/analytics/PlatformBreakdown.vue';
import StatCard from '@/components/analytics/StatCard.vue';
import { useTagStore } from '@/stores/tagStore';
import { useVideoStore } from '@/stores/videoStore';
import { formatFileSize, groupByWeek } from '@/utils/formatters';

const router = useRouter();
const tagStore = useTagStore();
const videoStore = useVideoStore();
const summary = ref(null);
const heatmap = ref([]);

const stats = computed(() => [
  { label: 'Total uploaded', value: summary.value?.totalUploaded ?? null, icon: FileVideo },
  { label: 'Posted', value: summary.value?.totalPosted ?? null, icon: CheckCircle2 },
  { label: 'Scheduled', value: summary.value?.totalScheduled ?? null, icon: CalendarClock },
  { label: 'Draft', value: summary.value?.totalDraft ?? null, icon: Layers },
  { label: 'This week', value: summary.value?.postsThisWeek ?? null, trend: summary.value?.weeklyTrend, icon: Flame },
  { label: 'Avg / week', value: summary.value?.avgPostsPerWeek ?? null, icon: BarChart3 },
  { label: 'Top platform', value: summary.value?.mostActivePlatform || 'None', icon: Hash },
  { label: 'Storage', value: summary.value ? formatFileSize(summary.value.totalStorageBytes) : null, icon: Database }
]);

const weekly = computed(() => {
  const rows = groupByWeek(heatmap.value, 12);
  const max = Math.max(...rows.map((row) => row.count), 1);
  return rows.map((row) => ({
    ...row,
    bucket: row.count === 0 ? 0 : Math.max(10, Math.round((row.count / max) * 10) * 10)
  }));
});

const tagMax = computed(() => Math.max(...tagStore.tags.map((tag) => Number(tag.count || 0)), 1));
const tagSize = (count) => `tag-size-${Math.min(5, Math.round((Number(count || 0) / tagMax.value) * 5))}`;

async function loadAnalytics() {
  const [summaryResponse, heatmapResponse] = await Promise.all([getSummary(), getHeatmap()]);
  summary.value = summaryResponse.data;
  heatmap.value = heatmapResponse.data;
  await tagStore.fetchTags();
}

function goToTag(tag) {
  videoStore.setFilters({ status: 'scheduled', tags: [tag.name] });
  router.push('/scheduled');
}

onMounted(() => {
  loadAnalytics().catch(console.error);
});
</script>

<template>
  <section class="space-y-5">
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <StatCard
        v-for="item in stats"
        :key="item.label"
        :label="item.label"
        :value="item.value"
        :trend="item.trend"
        :icon="item.icon"
      />
    </div>

    <PostingHeatmap :data="heatmap" />

    <div class="grid gap-5 lg:grid-cols-2">
      <PlatformBreakdown :breakdown="summary?.platformBreakdown || {}" />
      <section class="panel rounded-md p-4">
        <h2 class="mb-4 font-display text-xl font-semibold">Weekly posting</h2>
        <div class="flex h-56 items-end gap-2">
          <div v-for="week in weekly" :key="week.label" class="flex h-full flex-1 flex-col justify-end gap-2">
            <div
              class="rounded-t bg-accent transition-[height]"
              :class="[week.count ? `bar-h-${week.bucket}` : 'bar-h-0 border border-dashed border-border bg-transparent']"
              :title="`${week.count} posts`"
            />
            <span class="truncate text-center font-mono text-[10px] text-secondary">{{ week.label }}</span>
          </div>
        </div>
      </section>
    </div>

    <section class="panel rounded-md p-4">
      <h2 class="mb-4 font-display text-xl font-semibold">Tag cloud</h2>
      <div class="flex flex-wrap gap-2">
        <button
          v-for="tag in tagStore.tags"
          :key="tag.id"
          type="button"
          class="rounded-full border border-border bg-elevated px-3 py-1 text-accent transition-colors hover:border-accent hover:bg-accent hover:text-white"
          :class="tagSize(tag.count)"
          @click="goToTag(tag)"
        >
          {{ tag.name }} <span class="font-mono text-xs opacity-70">{{ tag.count }}</span>
        </button>
      </div>
    </section>
  </section>
</template>
