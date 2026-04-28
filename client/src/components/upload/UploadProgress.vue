<script setup>
defineProps({
  progress: { type: Number, default: 0 },
  complete: { type: Boolean, default: false }
});
</script>

<template>
  <progress class="upload-progress" :class="{ complete }" max="100" :value="Math.min(progress, 100)" />
</template>
