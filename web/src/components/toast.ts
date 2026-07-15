import { reactive } from 'vue'

export type ToastTone = 'success' | 'error' | 'info'

export interface ToastMessage {
  id: number
  message: string
  tone: ToastTone
}

export const toasts = reactive<ToastMessage[]>([])

export function showToast(message: string, tone: ToastTone = 'info'): void {
  const id = Date.now() + Math.floor(Math.random() * 1000)
  toasts.push({ id, message, tone })
  window.setTimeout(() => {
    const index = toasts.findIndex((toast) => toast.id === id)
    if (index >= 0) {
      toasts.splice(index, 1)
    }
  }, 3500)
}
