<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import {
  BarChart3,
  CalendarClock,
  CheckCircle2,
  Upload,
  Zap,
  ZapOff,
  X
} from 'lucide-vue-next';
import { api } from '@/api/videos';
import ActivityFeed from '@/components/activity/ActivityFeed.vue';

const props = defineProps({
  open: { type: Boolean, default: false }
});

const emit = defineEmits(['update:open']);
const route = useRoute();
const reachable = ref(false);
const expandedActivity = ref(false);
let timer = null;

const navItems = [
  { to: '/scheduled', label: 'Scheduled', icon: CalendarClock },
  { to: '/posted', label: 'Posted', icon: CheckCircle2 },
  { to: '/upload', label: 'Upload', icon: Upload },
  { to: '/analytics', label: 'Analytics', icon: BarChart3 }
];

const activeIndex = computed(() => Math.max(0, navItems.findIndex((item) => route.path.startsWith(item.to))));
const activeIndicatorClass = computed(() => `active-indicator-${activeIndex.value}`);

async function pollHealth() {
  try {
    await api.get('/api/health/n8n');
    reachable.value = true;
  } catch {
    reachable.value = false;
  }
}

onMounted(() => {
  pollHealth();
  timer = window.setInterval(pollHealth, 60000);
});

onUnmounted(() => {
  window.clearInterval(timer);
});
</script>

<template>
  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="props.open" class="fixed inset-0 z-50 md:hidden">
        <button class="absolute inset-0 bg-black/70" aria-label="Close navigation" @click="emit('update:open', false)" />
        <aside class="panel relative h-full w-[min(86vw,320px)] p-4">
          <div class="mb-6 flex items-center justify-between">
            <div>
              <p class="font-display text-lg font-semibold">V.C.M</p>
              <p class="text-xs text-secondary">Vertical Content Manager</p>
            </div>
            <button class="icon-button" type="button" aria-label="Close navigation" @click="emit('update:open', false)">
              <X class="h-4 w-4" :stroke-width="1.5" />
            </button>
          </div>
          <nav class="space-y-2">
            <RouterLink
              v-for="item in navItems"
              :key="item.to"
              :to="item.to"
              class="flex min-h-11 items-center gap-3 rounded-md px-3 text-sm text-secondary transition-colors hover:bg-white/10 hover:text-primary"
              :class="{ 'bg-accent/15 text-primary': route.path.startsWith(item.to) }"
              @click="emit('update:open', false)"
            >
              <component :is="item.icon" class="h-5 w-5" :stroke-width="1.5" />
              {{ item.label }}
            </RouterLink>
          </nav>
        </aside>
      </div>
    </Transition>
  </Teleport>

  <aside class="fixed inset-y-0 left-0 z-40 hidden border-r border-border bg-surface/95 backdrop-blur md:block md:w-16 lg:w-60">
    <div class="flex h-full flex-col p-3">
      <div class="mb-6 flex h-12 items-center gap-3 px-1">
        <div class="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-accent text-white">
          <Zap class="h-5 w-5" :stroke-width="1.5" />
        </div>
        <div class="hidden lg:block">
          <p class="font-display text-sm font-semibold leading-tight">V.C.M</p>
          <p class="text-xs text-secondary">Vertical Content Manager</p>
        </div>
      </div>

      <nav class="relative space-y-2">
        <span
          class="absolute left-0 h-11 w-[3px] rounded-full bg-accent transition-[top] ease-out"
          :class="activeIndicatorClass"
        />
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="group relative flex min-h-11 items-center gap-3 rounded-md px-3 text-sm text-secondary transition-colors hover:bg-white/10 hover:text-primary"
          :class="{ 'bg-accent/10 text-primary': route.path.startsWith(item.to) }"
        >
          <component :is="item.icon" class="h-5 w-5 shrink-0" :stroke-width="1.5" />
          <span class="hidden lg:inline">{{ item.label }}</span>
          <span class="pointer-events-none absolute left-14 rounded-md bg-elevated px-2 py-1 text-xs text-primary opacity-0 shadow-xl transition-opacity group-hover:opacity-100 lg:hidden">
            {{ item.label }}
          </span>
        </RouterLink>
      </nav>

      <div class="mt-auto space-y-3">
        <button
          class="flex w-full items-center gap-3 rounded-md border border-border bg-elevated/60 px-3 py-3 text-left transition-colors hover:bg-white/10"
          type="button"
          @click="expandedActivity = !expandedActivity"
        >
          <span
            class="h-2.5 w-2.5 rounded-full"
            :class="[reachable ? 'bg-green n8n-pulse' : 'bg-warm']"
          />
          <component :is="reachable ? Zap : ZapOff" class="h-4 w-4 text-secondary" :stroke-width="1.5" />
          <span class="hidden text-xs text-secondary lg:inline">{{ reachable ? 'n8n reachable' : 'n8n offline' }}</span>
        </button>

        <button
          class="group hidden min-h-[46px] w-full items-center justify-center rounded-md border border-accent bg-base/45 p-3 text-center text-accent transition-colors hover:bg-accent hover:text-black focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent focus-visible:ring-offset-2 focus-visible:ring-offset-base lg:flex"
          type="button"
          aria-label="Open activity feed"
          @click="expandedActivity = !expandedActivity"
        >
          <span class="text-xs font-semibold uppercase tracking-[0.16em]">Activity</span>
        </button>
      </div>
    </div>
  </aside>

  <Teleport to="body">
    <Transition name="drawer">
      <div v-if="expandedActivity" class="fixed inset-0 z-[70]">
        <button class="absolute inset-0 bg-black/70" aria-label="Close activity" @click="expandedActivity = false" />
        <section class="panel absolute bottom-0 right-0 top-0 w-full max-w-md p-4">
          <div class="mb-4 flex items-center justify-between">
            <h2 class="font-display text-xl font-semibold">Activity</h2>
            <button class="icon-button" type="button" aria-label="Close activity" @click="expandedActivity = false">
              <X class="h-4 w-4" :stroke-width="1.5" />
            </button>
          </div>
          <ActivityFeed />
        </section>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.drawer-enter-active,
.drawer-leave-active {
  transition: opacity var(--transition-slow);
}

.drawer-enter-from,
.drawer-leave-to {
  opacity: 0;
}

.drawer-enter-active .panel,
.drawer-leave-active .panel {
  transition: opacity var(--transition-slow), transform var(--transition-slow);
  will-change: opacity, transform;
}

.drawer-enter-from aside.panel,
.drawer-leave-to aside.panel {
  transform: translate3d(-100%, 0, 0);
}

.drawer-enter-from section.panel,
.drawer-leave-to section.panel {
  transform: translate3d(100%, 0, 0);
}
</style>
