import { defineStore } from 'pinia';

let toastId = 0;

export const useUiStore = defineStore('ui', {
  state: () => ({
    activeModal: null,
    toasts: [],
    isDraggingFile: false
  }),
  actions: {
    pushToast({ message, type = 'info', duration = 4000 }) {
      const id = toastId += 1;
      this.toasts.push({ id, message, type, duration });
      window.setTimeout(() => this.dismissToast(id), duration);
    },
    dismissToast(id) {
      this.toasts = this.toasts.filter((toast) => toast.id !== id);
    },
    setDragging(value) {
      this.isDraggingFile = value;
    },
    openModal(video) {
      this.activeModal = video;
    },
    closeModal() {
      this.activeModal = null;
    }
  }
});
