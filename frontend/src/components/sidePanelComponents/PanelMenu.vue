<template>
	<div v-if="IsAuth" class="flex flex-col h-full w-full p-2 gap-1">
		<button class="text-left px-3 py-2 rounded bg-accent text-surface" @click="startNew">+ New page</button>
		<div class="flex-1 overflow-auto">
			<button v-for="p in state.pages" :key="p.PageName"
				class="block w-full text-left px-3 py-2 rounded hover:bg-surface-2 text-fg"
				:class="{ 'bg-surface-2 font-semibold': state.selected?.PageName === p.PageName }"
				@click="select(p.PageName)">
				{{ p.PageName }}
			</button>
		</div>
	</div>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue'
import { usePanelState, refresh, select, startNew } from '@/global/panelState'
import UserService from '@/service/User.service'

const state = usePanelState()
const IsAuth = UserService.IsLogin

function load() {
	if (IsAuth.value) refresh()
}
onMounted(load)
watch(IsAuth, load) // refresh once the login resolves
</script>
