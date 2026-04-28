<script setup>
import { Inbox } from 'lucide-vue-next';

defineProps({
  icon: { type: [Object, Function], default: Inbox },
  title: { type: String, required: true },
  description: { type: String, required: true },
  actionLabel: { type: String, default: '' },
  actionFn: { type: Function, default: null }
});
</script>

<template>
  <div class="flex min-h-[320px] flex-col items-center justify-center rounded-md border border-dashed border-border bg-surface/55 p-8 text-center">
    <component :is="icon" class="mb-4 h-12 w-12 text-secondary" :stroke-width="1.5" />
    <h2 class="font-display text-xl font-semibold">{{ title }}</h2>
    <p class="mt-2 max-w-md text-sm text-secondary">{{ description }}</p>
    <button v-if="actionLabel && actionFn" type="button" class="btn-primary mt-5" @click="actionFn">
      {{ actionLabel }}
    </button>
  </div>
</template>
