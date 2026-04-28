<script setup>
import { X } from 'lucide-vue-next';

defineProps({
  title: { type: String, required: true },
  message: { type: String, required: true },
  confirmLabel: { type: String, default: 'Confirm' },
  destructive: { type: Boolean, default: false }
});

defineEmits(['confirm', 'cancel']);
</script>

<template>
  <Teleport to="body">
    <Transition name="dialog" appear>
      <div class="fixed inset-0 z-[90] flex items-center justify-center bg-black/70 p-4">
        <section class="panel w-full max-w-md rounded-md p-5">
          <div class="mb-4 flex items-start justify-between gap-4">
            <div>
              <h2 class="font-display text-xl font-semibold">{{ title }}</h2>
              <p class="mt-2 text-sm text-secondary">{{ message }}</p>
            </div>
            <button class="icon-button h-9 w-9" type="button" aria-label="Cancel" @click="$emit('cancel')">
              <X class="h-4 w-4" :stroke-width="1.5" />
            </button>
          </div>
          <div class="flex justify-end gap-3">
            <button type="button" class="btn-secondary" @click="$emit('cancel')">Cancel</button>
            <button type="button" :class="destructive ? 'btn-danger' : 'btn-primary'" @click="$emit('confirm')">
              {{ confirmLabel }}
            </button>
          </div>
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.dialog-enter-active,
.dialog-leave-active {
  transition: opacity var(--transition-base);
}

.dialog-enter-active .panel,
.dialog-leave-active .panel {
  transition: opacity var(--transition-base), transform var(--transition-base);
  will-change: opacity, transform;
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}

.dialog-enter-from .panel,
.dialog-leave-to .panel {
  opacity: 0;
  transform: translate3d(0, 10px, 0) scale(0.98);
}
</style>
