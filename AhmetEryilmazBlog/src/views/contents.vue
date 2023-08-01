<template>
	<div class="py-16 pl-8">
		<div class="rounded border-2 border-white h-full w-full overscroll-none">
			{{ returnedHTML }}
		</div>
	</div>
</template>

<style scoped>
* {
	overflow-y: auto;
}
</style>

<script setup lang="ts">
import { ref } from 'vue'
const returnedHTML = ref<any>('selam canim');
</script>