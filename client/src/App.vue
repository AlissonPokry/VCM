<script setup>
import { storeToRefs } from 'pinia';
import { CheckCircle2, Info, X, AlertTriangle } from 'lucide-vue-next';
import AppShell from '@/components/layout/AppShell.vue';
import VideoModal from '@/components/video/VideoModal.vue';
import { useUiStore } from '@/stores/uiStore';
import { useVideoStore } from '@/stores/videoStore';

const uiStore = useUiStore();
const videoStore = useVideoStore();
const { toasts } = storeToRefs(uiStore);

const toastIcons = {
  success: CheckCircle2,
  error: AlertTriangle,
  info: Info
};
</script>

<template>
  <AppShell>
    <RouterView v-slot="{ Component }">
      <Transition name="route" mode="out-in">
        <component :is="Component" />
      </Transition>
    </RouterView>
  </AppShell>

  <VideoModal
    v-if="videoStore.selectedVideo"
    :video="videoStore.selectedVideo"
    @close="videoStore.selectedVideo = null"
  />

  <TransitionGroup
    name="toast"
    tag="div"
    class="fixed bottom-4 right-4 z-[80] flex w-[min(92vw,360px)] flex-col gap-2"
  >
    <button
      v-for="toast in toasts"
      :key="toast.id"
      class="panel group overflow-hidden rounded-md p-4 text-left transition-transform active:scale-[.98]"
      @click="uiStore.dismissToast(toast.id)"
    >
      <div class="flex items-start gap-3">
        <component
          :is="toastIcons[toast.type] || Info"
          class="mt-0.5 h-4 w-4"
          :class="toast.type === 'error' ? 'text-warm' : toast.type === 'success' ? 'text-green' : 'text-accent'"
          :stroke-width="1.5"
        />
        <p class="text-sm text-primary">{{ toast.message }}</p>
        <X class="ml-auto h-4 w-4 text-secondary transition-colors group-hover:text-primary" :stroke-width="1.5" />
      </div>
      <span class="mt-3 block h-0.5 origin-left bg-accent" />
    </button>
  </TransitionGroup>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: opacity var(--transition-fast), transform var(--transition-spring);
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateX(110%);
}

.toast-move {
  transition: transform var(--transition-base);
}
</style>
