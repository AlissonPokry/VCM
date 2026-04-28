import { createApp } from 'vue';
import { createPinia } from 'pinia';
import router from './router';
import App from './App.vue';
import './assets/styles/main.css';
import { useActivityStore } from './stores/activityStore';
import { useTagStore } from './stores/tagStore';

const app = createApp(App);
const pinia = createPinia();

app.use(pinia);
app.use(router);

useTagStore().fetchTags().catch(console.error);
useActivityStore().fetchActivity().catch(console.error);

app.mount('#app');
