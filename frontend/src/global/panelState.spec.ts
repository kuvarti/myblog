import { describe, it, expect, vi, beforeEach } from 'vitest'

vi.mock('@/service/Panel.service', () => ({
	default: { listPages: vi.fn(), getPageDetail: vi.fn() },
}))

import PanelService from '@/service/Panel.service'
import { usePanelState, refresh, select, startNew } from '@/global/panelState'

describe('panelState', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		const s = usePanelState()
		s.pages = []
		s.selected = null
		s.dirty = false
		s.isNew = false
	})

	it('refresh populates pages from the service', async () => {
		;(PanelService.listPages as any).mockResolvedValue([{ PageName: 'A', ViewType: '' }])
		await refresh()
		expect(usePanelState().pages).toEqual([{ PageName: 'A', ViewType: '' }])
	})

	it('select loads a page and clears dirty/new', async () => {
		;(PanelService.getPageDetail as any).mockResolvedValue({ PageName: 'A', Source: 'x', ViewType: '', Menu: null })
		usePanelState().dirty = true
		await select('A')
		const s = usePanelState()
		expect(s.selected?.PageName).toBe('A')
		expect(s.dirty).toBe(false)
		expect(s.isNew).toBe(false)
	})

	it('startNew creates an empty editable page', () => {
		startNew()
		const s = usePanelState()
		expect(s.isNew).toBe(true)
		expect(s.selected?.PageName).toBe('')
	})
})
