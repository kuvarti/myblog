<template>
	<div class="py-16 pl-8">
		<div class="">
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
import { computed } from 'vue';
import { useStore } from 'vuex';

let GlobalStore = useStore()
const returnedHTML = computed(() => {
	return GlobalStore.getters.GetScreenLevel
})
</script>
