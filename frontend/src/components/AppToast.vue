<template>
	<teleport to="body">
		<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2 items-end">
			<div v-for="t in toasts.toasts" :key="t.id"
				class="px-4 py-2 rounded shadow border text-sm text-fg bg-surface-2"
				:class="t.kind === 'error' ? 'border-red-500' : 'border-accent'"
				role="status">
				{{ t.message }}
			</div>
		</div>
	</teleport>
</template>

<script setup lang="ts">
import { useToasts } from '@/global/notify'

const toasts = useToasts()
</script>
