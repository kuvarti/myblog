import { reactive } from 'vue'
import PanelService from '@/service/Panel.service'
import { moveItem } from '@/components/sidePanelComponents/menuReorder'
import { notify } from '@/global/notify'
import { refreshMenu } from '@/global/menuRefresh'
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

export async function setVisibility(name: string, visible: boolean): Promise<void> {
	const page = state.pages.find((p) => p.PageName === name)
	if (page) page.Visible = visible // optimistic
	try {
		await PanelService.setVisibility(name, visible)
		notify(`${name} is now ${visible ? 'visible in' : 'hidden from'} the menu`)
		refreshMenu() // the public side menu refetches
	} catch {
		notify(`Failed to update ${name}`, 'error')
		await refresh() // resync on failure
	}
}

export function startNew(): void {
	state.selected = { PageName: '', Path: '', Source: '', ViewType: '', Menu: { Name: '', Caption: '' } }
	state.isNew = true
	state.dirty = false
}
