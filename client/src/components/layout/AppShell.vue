<script setup>
import { ref } from 'vue';
import AppHeader from './AppHeader.vue';
import AppSidebar from './AppSidebar.vue';

const sidebarOpen = ref(false);
</script>

<template>
  <div class="min-h-screen bg-base text-primary">
    <AppSidebar v-model:open="sidebarOpen" />
    <div class="min-h-screen transition-[padding] duration-150 md:pl-16 lg:pl-60">
      <AppHeader @toggle-sidebar="sidebarOpen = true" />
      <main class="mx-auto w-full max-w-[1520px] px-4 py-5 sm:px-6 lg:px-8">
        <slot />
      </main>
    </div>
  </div>
</template>
