<template>
	<button
		type="button"
		class="text-sm font-medium text-fg bg-surface border border-border rounded-full px-3 py-1 hover:border-accent"
		@click="onToggle"
	>
		{{ current === 'dark' ? '☾ Dark' : '☀ Light' }}
	</button>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { getTheme, toggleTheme, type Theme } from '@/global/theme'

const current = ref<Theme>(getTheme())
function onToggle() {
	current.value = toggleTheme()
}
</script>
