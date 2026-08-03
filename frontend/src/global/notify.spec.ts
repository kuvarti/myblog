import { describe, it, expect, beforeEach, vi } from 'vitest'
import { notify, dismiss, useToasts } from './notify'

describe('notify', () => {
	beforeEach(() => {
		useToasts().toasts.splice(0) // clear the shared singleton between tests
	})

	it('adds a toast with the given message and default success kind', () => {
		notify('saved', 'success', 0) // durationMs 0 → no auto-dismiss
		const { toasts } = useToasts()
		expect(toasts).toHaveLength(1)
		expect(toasts[0]).toMatchObject({ message: 'saved', kind: 'success' })
	})

	it('records the error kind', () => {
		notify('boom', 'error', 0)
		expect(useToasts().toasts[0].kind).toBe('error')
	})

	it('dismiss removes the toast by id', () => {
		const id = notify('bye', 'success', 0)
		dismiss(id)
		expect(useToasts().toasts).toHaveLength(0)
	})

	it('auto-dismisses after the duration', () => {
		vi.useFakeTimers()
		notify('temp', 'success', 2500)
		expect(useToasts().toasts).toHaveLength(1)
		vi.advanceTimersByTime(2500)
		expect(useToasts().toasts).toHaveLength(0)
		vi.useRealTimers()
	})
})
