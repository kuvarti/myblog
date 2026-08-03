import { reactive } from 'vue'
import PanelService from '@/service/Panel.service'
import { moveItem } from '@/components/sidePanelComponents/menuReorder'
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

export async function reorder(from: number, to: number): Promise<void> {
	const next = moveItem(state.pages, from, to)
	state.pages = next // optimistic; the sorted list persists below
	await PanelService.reorderPages(next.map((p) => p.PageName))
}

export function startNew(): void {
	state.selected = { PageName: '', Source: '', ViewType: '', Menu: { Name: '', Caption: '', Path: '' } }
	state.isNew = true
	state.dirty = false
}
