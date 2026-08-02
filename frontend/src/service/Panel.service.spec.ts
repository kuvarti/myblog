import { describe, it, expect, vi, beforeEach } from 'vitest'

// Shared axios spies (both UserService and PanelService call axios.create).
// Default resolutions keep UserService's AmIAuth (fired in its constructor) happy.
// vi.hoisted keeps them available inside the hoisted vi.mock factory.
const { getMock, postMock, putMock, deleteMock } = vi.hoisted(() => ({
	getMock: vi.fn(() => Promise.resolve({ data: [] })),
	postMock: vi.fn(() => Promise.resolve({ data: {} })),
	putMock: vi.fn(() => Promise.resolve({ data: {} })),
	deleteMock: vi.fn(() => Promise.resolve({ data: {} })),
}))
vi.mock('axios', () => ({
	default: { create: () => ({ get: getMock, post: postMock, put: putMock, delete: deleteMock }) },
}))

// Mock UserService so its constructor side effects (localStorage read + AmIAuth
// request) don't run at import time; Panel.service only reads UserService.IsLogin.
vi.mock('@/service/User.service', () => ({
	default: { IsLogin: { value: false } },
}))

// jsdom runs on an opaque origin where localStorage is unavailable — stub it.
const store: Record<string, string> = {}
vi.stubGlobal('localStorage', {
	getItem: (k: string) => (k in store ? store[k] : null),
	setItem: (k: string, v: string) => { store[k] = v },
	clear: () => { for (const k in store) delete store[k] },
})

import PanelService from '@/service/Panel.service'
import UserService from '@/service/User.service'

describe('Panel.service', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		store['AuthToken'] = 'tok123'
	})

	it('attaches the bearer token from local storage', async () => {
		getMock.mockResolvedValueOnce({ data: [] })
		await PanelService.listPages()
		expect(getMock).toHaveBeenCalledWith('/Pages', {
			headers: { Authorization: 'Bearer tok123' },
		})
	})

	it('flips IsLogin to false on a 401', async () => {
		UserService.IsLogin.value = true
		getMock.mockRejectedValueOnce({ response: { status: 401 } })
		await PanelService.listPages()
		expect(UserService.IsLogin.value).toBe(false)
	})
})
