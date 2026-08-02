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
					<label class="text-sm text-muted mb-1">View type</label>
					<InputText v-model="state.selected.ViewType"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Menu caption</label>
					<InputText v-model="menu.Caption"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
				<div class="flex flex-col">
					<label class="text-sm text-muted mb-1">Menu path</label>
					<InputText v-model="menu.Path"
						class="bg-surface-2 border border-border text-fg rounded p-2" />
				</div>
			</div>

			<div class="flex flex-1 gap-4 min-h-0">
				<textarea v-model="source" @input="onInput"
					class="flex-1 bg-surface-2 border border-border text-fg rounded p-3 font-mono resize-none"
					placeholder="Write Markdown here..."></textarea>
				<div class="flex-1 overflow-auto border border-border rounded p-3">
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
				<span v-if="error" class="text-sm" style="color:#c0392b">{{ error }}</span>
			</div>
		</template>
	</div>
</template>

<script setup lang="ts">
import InputText from 'primevue/inputtext'
import Button from 'primevue/button'
import { ref, watch } from 'vue'
import { usePanelState, refresh, select } from '@/global/panelState'
import PanelService from '@/service/Panel.service'
import type { MenuBinding } from '@/types/PanelModels'

const state = usePanelState()
const source = ref('')
const previewHtml = ref('')
const menu = ref<MenuBinding>({ Name: '', Caption: '', Path: '' })
const confirming = ref(false)
const error = ref('')

let debounce: ReturnType<typeof setTimeout> | undefined

watch(() => state.selected, (sel) => {
	source.value = sel?.Source ?? ''
	menu.value = sel?.Menu ?? { Name: '', Caption: '', Path: '' }
	confirming.value = false
	error.value = ''
	runPreview()
}, { immediate: true })

function onInput() {
	state.dirty = true
	if (debounce) clearTimeout(debounce)
	debounce = setTimeout(runPreview, 400)
}

async function runPreview() {
	previewHtml.value = await PanelService.preview(source.value)
}

async function save() {
	if (!state.selected) return
	error.value = ''
	const hasMenu = !!(menu.value.Caption || menu.value.Path)
	try {
		if (state.isNew) {
			await PanelService.createPage({
				PageName: state.selected.PageName,
				Source: source.value,
				ViewType: state.selected.ViewType,
				Menu: hasMenu ? menu.value : null,
			})
		} else {
			await PanelService.updatePage(state.selected.PageName, {
				Source: source.value,
				ViewType: state.selected.ViewType,
				Menu: hasMenu ? menu.value : null,
			})
		}
		const name = state.selected.PageName
		state.dirty = false
		await refresh()
		await select(name)
	} catch (e: any) {
		error.value = e?.response?.status === 409
			? 'A page with that name already exists.'
			: 'Save failed.'
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
	} catch {
		error.value = 'Delete failed.'
	} finally {
		confirming.value = false
	}
}
</script>
