import { reactive } from 'vue'
import PanelService from '@/service/Panel.service'
import type { PanelPageSummary, PanelPageDetail } from '@/types/PanelModels'

interface PanelState {
	pages: PanelPageSummary[]
	selected: PanelPageDetail | null
	isNew: boolean
	dirty: boolean
}

const state = reactive<PanelState>({ pages: [], selected: null, isNew: false, dirty: false })

export function usePanelState() {
	return state
}

export async function refresh(): Promise<void> {
	state.pages = await PanelService.listPages()
}

export async function select(name: string): Promise<void> {
	state.selected = await PanelService.getPageDetail(name)
	state.isNew = false
	state.dirty = false
}

export function startNew(): void {
	state.selected = { PageName: '', Source: '', ViewType: '', Menu: { Name: '', Caption: '', Path: '' } }
	state.isNew = true
	state.dirty = false
}
