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
				<div class="ml-auto flex items-center gap-2">
					<span v-if="state.dirty" class="text-muted text-sm">unsaved changes</span>
					<Button label="Save" class="bg-accent text-surface rounded px-4 py-2" @click="save" />
					<button type="button" title="Editor guide" @click="showHelp = true"
						class="border border-border text-muted hover:text-fg rounded px-3 py-2">?</button>
				</div>
			</div>

			<div class="flex flex-wrap gap-3 items-end">
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
				<div v-if="!state.isNew" class="ml-auto flex items-center gap-2">
					<Button v-if="!confirming" label="Delete Page"
						class="border border-border text-fg rounded px-4 py-2" @click="confirming = true" />
					<template v-else>
						<span class="text-fg text-sm">Really delete?</span>
						<Button label="Yes" class="bg-accent text-surface rounded px-3 py-1" @click="doDelete" />
						<Button label="No" class="border border-border text-fg rounded px-3 py-1" @click="confirming = false" />
					</template>
				</div>
			</div>

			<div class="flex flex-1 gap-4 min-h-0">
				<textarea ref="editorEl" v-model="source" @input="onInput" @scroll="onEditorScroll"
					@keydown.tab.prevent="onEditorTab"
					class="flex-1 bg-surface-2 border border-border text-fg rounded p-3 font-mono resize-none"
					placeholder="Write Markdown here..."></textarea>
				<div ref="previewEl" @scroll="onPreviewScroll"
					class="flex-1 overflow-auto border border-border rounded p-3">
					<div class="content" v-html="previewHtml"></div>
				</div>
			</div>

			<teleport to="body">
				<div v-if="showHelp" class="fixed inset-0 z-50 flex items-center justify-center p-4">
					<div class="absolute inset-0 bg-black/50" @click="showHelp = false"></div>
					<div class="relative bg-surface border border-border rounded-lg w-full max-w-lg max-h-[80vh] overflow-auto p-6 text-fg">
						<div class="flex items-center justify-between mb-4">
							<h3 class="text-lg font-semibold">Editor guide</h3>
							<button type="button" class="text-muted hover:text-fg text-xl leading-none" @click="showHelp = false">✕</button>
						</div>
						<div class="text-sm text-muted space-y-3">
							<p><strong class="text-fg">Format.</strong> Write Markdown. Any line containing <code>&lt;</code> is emitted as raw HTML.</p>
							<p><strong class="text-fg">Cards.</strong> Put <code>&lt;card path="/some-path"&gt;</code> on its own line to render that page as a linked card — its title, summary and image are pulled from that page.</p>
							<p><strong class="text-fg">View type.</strong> <code>PlainHTML</code> is a normal page. <code>List</code> auto-lists every page sharing a tag with this page's <em>List tags</em>, as a card grid.</p>
							<p><strong class="text-fg">Tags.</strong> Comma-separated. A page must be tagged to appear in a List.</p>
							<p><strong class="text-fg">Card summary / image.</strong> Optional overrides — leave blank to auto-use the page's first paragraph / first image.</p>
							<p><strong class="text-fg">Tab.</strong> Inserts an indent; <em>Shift+Tab</em> outdents.</p>
						</div>
					</div>
				</div>
			</teleport>
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
import { applyTab } from '@/components/panelComponents/editorTab'
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
const showHelp = ref(false)

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

// Tab / Shift+Tab indent the editor text instead of moving focus.
function onEditorTab(e: KeyboardEvent) {
	const ta = editorEl.value
	if (!ta) return
	const r = applyTab(source.value, ta.selectionStart, ta.selectionEnd, e.shiftKey)
	source.value = r.value
	onInput()
	nextTick(() => {
		ta.selectionStart = r.selStart
		ta.selectionEnd = r.selEnd
	})
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
