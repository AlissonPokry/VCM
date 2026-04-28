import { reactive, ref } from 'vue';
import { useVideoStore } from '@/stores/videoStore';
import { toIsoFromInput } from '@/utils/formatters';

export function useVideoUpload() {
  const videoStore = useVideoStore();
  const files = ref([]);
  const progress = reactive({});
  const uploading = ref(false);

  function addFiles(fileList) {
    const mapped = Array.from(fileList)
      .filter((file) => file.type.startsWith('video/'))
      .map((file) => ({
        id: `${file.name}-${file.lastModified}-${file.size}`,
        file,
        title: file.name.replace(/\.[^/.]+$/, ''),
        description: '',
        platform: 'instagram',
        scheduled_at: '',
        tags: []
      }));
    files.value = [...files.value, ...mapped];
  }

  function removeFile(id) {
    files.value = files.value.filter((item) => item.id !== id);
    delete progress[id];
  }

  async function uploadAll() {
    uploading.value = true;
    try {
      for (const item of files.value) {
        const form = new FormData();
        form.append('file', item.file);
        form.append('title', item.title);
        form.append('description', item.description);
        form.append('platform', item.platform);
        form.append('scheduled_at', toIsoFromInput(item.scheduled_at) || '');
        form.append('tags', JSON.stringify(item.tags));
        await videoStore.uploadVideo(form, (event) => {
          progress[item.id] = event.total ? Math.round((event.loaded / event.total) * 100) : 0;
        });
        progress[item.id] = 100;
        window.setTimeout(() => removeFile(item.id), 150);
      }
    } finally {
      uploading.value = false;
    }
  }

  return { files, addFiles, removeFile, uploadAll, progress, uploading };
}
