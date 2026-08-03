<template>
	<div v-if="IsAuth" class="flex flex-col h-full w-full p-2 gap-1">
		<button class="text-left px-3 py-2 rounded bg-accent text-surface" @click="startNew">+ New page</button>
		<div class="flex-1 overflow-auto">
			<button v-for="(p, i) in state.pages" :key="p.PageName"
				draggable="true"
				@dragstart="onDragStart(i)"
				@dragover.prevent="onDragOver(i)"
				@drop.prevent="onDrop(i)"
				@dragend="onDragEnd"
				@click="select(p.PageName)"
				class="flex items-center gap-2 w-full text-left px-3 py-2 rounded hover:bg-surface-2 text-fg cursor-move"
				:class="{
					'bg-surface-2 font-semibold': state.selected?.PageName === p.PageName,
					'opacity-40': dragIndex === i,
					'border-t-2 border-accent': overIndex === i && dragIndex !== null && dragIndex !== i,
				}">
				<v-icon name="hi-menu" class="text-muted shrink-0" aria-hidden="true" />
				<span>{{ p.PageName }}</span>
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, watch, ref } from 'vue'
import { usePanelState, refresh, select, startNew, reorder } from '@/global/panelState'
import UserService from '@/service/User.service'

const state = usePanelState()
const IsAuth = UserService.IsLogin

const dragIndex = ref<number | null>(null)
const overIndex = ref<number | null>(null)

function onDragStart(i: number) {
	dragIndex.value = i
}
function onDragOver(i: number) {
	overIndex.value = i
}
function onDrop(i: number) {
	if (dragIndex.value !== null && dragIndex.value !== i) {
		reorder(dragIndex.value, i)
	}
	onDragEnd()
}
function onDragEnd() {
	dragIndex.value = null
	overIndex.value = null
}

function load() {
	if (IsAuth.value) refresh()
}
onMounted(load)
watch(IsAuth, load) // refresh once the login resolves
</script>
