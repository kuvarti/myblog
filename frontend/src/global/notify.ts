import { reactive } from 'vue'

// Minimal app-wide toast notifications. A reactive list of active toasts is
// rendered by AppToast.vue; notify() adds one and auto-dismisses it after a
// short delay. Kept small and framework-light to match the codebase style.
export type ToastKind = 'success' | 'error'

export interface Toast {
	id: number
	message: string
	kind: ToastKind
}

const state = reactive<{ toasts: Toast[] }>({ toasts: [] })
let seq = 0

export function useToasts() {
	return state
}

export function notify(message: string, kind: ToastKind = 'success', durationMs = 2500): number {
	const id = ++seq
	state.toasts.push({ id, message, kind })
	if (durationMs > 0) {
		setTimeout(() => dismiss(id), durationMs)
	}
	return id
}

export function dismiss(id: number): void {
	const i = state.toasts.findIndex((t) => t.id === id)
	if (i !== -1) state.toasts.splice(i, 1)
}
