import { computed } from 'vue';
import { thumbnailUrl } from '@/api/videos';

export function useThumbnail(videoRef) {
  return computed(() => {
    const video = videoRef?.value || videoRef;
    return video?.thumbnail ? thumbnailUrl(video.id) : null;
  });
}
