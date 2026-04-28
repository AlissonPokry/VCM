<script setup>
import DropZone from '@/components/upload/DropZone.vue';
import ActivityFeed from '@/components/activity/ActivityFeed.vue';
</script>

<template>
  <section class="grid gap-5 xl:grid-cols-[1fr_360px]">
    <div>
      <DropZone />
    </div>
    <aside class="panel max-h-[720px] rounded-md p-4">
      <h2 class="mb-4 font-display text-xl font-semibold">Recent activity</h2>
      <ActivityFeed compact />
    </aside>
  </section>
</template>
