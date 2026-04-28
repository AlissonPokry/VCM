import { createRouter, createWebHistory } from 'vue-router';

const routes = [
  { path: '/', redirect: '/scheduled' },
  { path: '/scheduled', name: 'Scheduled', component: () => import('@/views/ScheduledView.vue') },
  { path: '/posted', name: 'Posted', component: () => import('@/views/PostedView.vue') },
  { path: '/upload', name: 'Upload', component: () => import('@/views/UploadView.vue') },
  { path: '/analytics', name: 'Analytics', component: () => import('@/views/AnalyticsView.vue') }
];

export default createRouter({
  history: createWebHistory(),
  routes
});
