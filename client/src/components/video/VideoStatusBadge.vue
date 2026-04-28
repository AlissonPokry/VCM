<script setup>
import { computed } from 'vue';
import { STATUSES } from '@/constants';

const props = defineProps({
  status: { type: String, required: true }
});

const statusMeta = computed(() => STATUSES.find((item) => item.value === props.status) || STATUSES[0]);
</script>

<template>
  <span class="rounded-full px-2.5 py-1 text-xs font-medium" :class="statusMeta.classes">
    {{ statusMeta.label }}
  </span>
</template>
