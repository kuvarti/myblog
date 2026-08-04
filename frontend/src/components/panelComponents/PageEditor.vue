<template>
	<div class="flex flex-col h-full p-4 gap-3">
		<div v-if="!state.selected" class="text-muted">Select a page or create a new one.</div>
		<template v-else>
			<div class="flex flex-wrap gap-3 items-end">
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Page name</label>
					<InputText v-model="state.selected.PageName" :disabled="!state.isNew"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Page path</label>
					<InputText v-model="state.selected.Path"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">View type</label>
					<select v-model="state.selected.ViewType"
						class="bg-surface-2 border border-border text-fg rounded p-2 w-44">
						<option value="PlainHTML">PlainHTML</option>
						<option value="List">List</option>
					</select>
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Tags</label>
					<InputText v-model="tagsInput" placeholder="blog, go"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div v-if="state.selected.ViewType === 'List'" class="flex flex-col">
					<label class="text-sm text-muted mb-1">List tags</label>
					<InputText v-model="listTagsInput" placeholder="blog"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Card summary (override)</label>
					<Textarea v-model="state.selected.Summary" rows="2" autoResize
						class="bg-surface-2 border border-border text-fg rounded p-2 w-96 resize-y" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Card image (override)</label>
					<InputText v-model="state.selected.Image"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Menu caption</label>
					<InputText v-model="menu.Caption"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
			</div>

			<div class="flex flex-1 gap-4 min-h-0">
				<textarea ref="editorEl" v-model="source" @input="onInput" @scroll="onEditorScroll"
					class="flex-1 bg-surface-2 border border-border text-fg rounded p-3 font-mono resize-none"
					placeholder="Write Markdown here..."></textarea>
				<div ref="previewEl" @scroll="onPreviewScroll"
					class="flex-1 overflow-auto border border-border rounded p-3">
					<div class="content" v-html="previewHtml"></div>
				</div>
			</div>

			<div class="flex items-center gap-3">
				<Button label="Save" class="bg-accent text-surface rounded px-4 py-2" @click="save" />
				<template v-if="!state.isNew">
					<Button v-if="!confirming" label="Delete"
						class="border border-border text-fg rounded px-4 py-2" @click="confirming = true" />
					<template v-else>
						<span class="text-fg text-sm">Really delete?</span>
						<Button label="Yes" class="bg-accent text-surface rounded px-3 py-1" @click="doDelete" />
						<Button label="No" class="border border-border text-fg rounded px-3 py-1" @click="confirming = false" />
					</template>
				</template>
				<span v-if="state.dirty" class="text-muted text-sm">unsaved changes</span>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import Textarea from 'primevue/textarea'
import { ref, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { usePanelState, refresh, select } from '@/global/panelState'
import { notify } from '@/global/notify'
import { refreshMenu } from '@/global/menuRefresh'
import PanelService from '@/service/Panel.service'
import type { MenuBinding } from '@/types/PanelModels'
import { parseTags, formatTags } from '@/components/panelComponents/tags'
import {
	splitBlocks, interpolateScroll, buildAnchorPair,
	measureEditorTops, measurePreviewTops,
} from '@/components/panelComponents/scrollSync'

const state = usePanelState()
const source = ref('')
const tagsInput = ref('')
const listTagsInput = ref('')
const previewHtml = ref('')
const menu = ref<MenuBinding>({ Name: '', Caption: '' })
const confirming = ref(false)

const editorEl = ref<HTMLTextAreaElement>()
const previewEl = ref<HTMLElement>()

let debounce: ReturnType<typeof setTimeout> | undefined

// --- Scroll sync: cached anchor offsets + a rAF lock against the feedback loop.
let editorTops: number[] = [0, 0]
let previewTops: number[] = [0, 0]
let syncing = false

function measureAnchors() {
	const ta = editorEl.value
	const pv = previewEl.value
	if (!ta || !pv) return
	const starts = splitBlocks(source.value)
	const pair = buildAnchorPair(
		measureEditorTops(ta, starts),
		measurePreviewTops(pv),
		ta.scrollHeight - ta.clientHeight,
		pv.scrollHeight - pv.clientHeight,
	)
	editorTops = pair.editorTops
	previewTops = pair.previewTops
}

function onEditorScroll() {
	if (syncing) return
	const ta = editorEl.value
	const pv = previewEl.value
	if (!ta || !pv) return
	syncing = true
	pv.scrollTop = interpolateScroll(editorTops, previewTops, ta.scrollTop)
	requestAnimationFrame(() => { syncing = false })
}

function onPreviewScroll() {
	if (syncing) return
	const ta = editorEl.value
	const pv = previewEl.value
	if (!ta || !pv) return
	syncing = true
	ta.scrollTop = interpolateScroll(previewTops, editorTops, pv.scrollTop)
	requestAnimationFrame(() => { syncing = false })
}

let ro: ResizeObserver | undefined
onMounted(() => {
	ro = new ResizeObserver(() => measureAnchors())
	if (editorEl.value) ro.observe(editorEl.value)
	if (previewEl.value) ro.observe(previewEl.value)
})
onBeforeUnmount(() => ro?.disconnect())

watch(() => state.selected, (sel) => {
	source.value = sel?.Source ?? ''
	menu.value = sel?.Menu ?? { Name: '', Caption: '' }
	confirming.value = false
	tagsInput.value = formatTags(sel?.Tags ?? [])
	listTagsInput.value = formatTags(sel?.ListTags ?? [])
	runPreview()
}, { immediate: true })

function onInput() {
	state.dirty = true
	if (debounce) clearTimeout(debounce)
	debounce = setTimeout(runPreview, 400)
}

async function runPreview() {
	previewHtml.value = await PanelService.preview(source.value)
	// Measure after the new preview DOM is patched in; the panes may not exist
	// yet on the first render (no page selected), so measureAnchors() no-ops.
	await nextTick()
	if (ro && editorEl.value) ro.observe(editorEl.value)
	if (ro && previewEl.value) ro.observe(previewEl.value)
	measureAnchors()
}

async function save() {
	if (!state.selected) return
	const isNew = state.isNew
	const name = state.selected.PageName
	const hasMenu = !!menu.value.Caption
	try {
		const fields = {
			Path: state.selected.Path,
			Source: source.value,
			ViewType: state.selected.ViewType,
			Tags: parseTags(tagsInput.value),
			Summary: state.selected.Summary ?? '',
			Image: state.selected.Image ?? '',
			ListTags: parseTags(listTagsInput.value),
			Menu: hasMenu ? menu.value : null,
		}
		if (isNew) {
			await PanelService.createPage({ PageName: name, ...fields })
		} else {
			await PanelService.updatePage(name, fields)
		}
		state.dirty = false
		await refresh()
		await select(name)
		notify(`${name} ${isNew ? 'created' : 'saved'}`)
		refreshMenu() // the public side menu refetches
	} catch (e: any) {
		notify(
			e?.response?.status === 409
				? 'A page with that name already exists.'
				: e?.response?.status === 422
					? 'That path is reserved or already used by another page.'
					: e?.response?.status === 400
						? 'Path must start with "/".'
						: 'Save failed.',
			'error',
		)
	}
}

async function doDelete() {
	if (!state.selected) return
	const name = state.selected.PageName
	try {
		await PanelService.deletePage(name)
		state.selected = null
		state.dirty = false
		await refresh()
		notify(`${name} deleted`)
		refreshMenu() // the public side menu refetches
	} catch {
		notify('Delete failed.', 'error')
	} finally {
		confirming.value = false
	}
}
</script>
