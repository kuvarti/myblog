<template>
	<div class="py-4 pl-4">
		<div class="">
			{{ returnedHTML }}
			<p> hy gidi hey gidi hey </p>
		</div>
	</div>
</template>

<style scoped>
* {
	overflow-y: auto;
}
</style>

<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';

let GlobalStore = useStore()
const returnedHTML = computed(() => {
	return GlobalStore.getters.GetActivePage
})
</script>
