<template>
	<div id="sliding" :class="ActiveClass">
	</div>
</template>

<style>
</style>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useStore } from 'vuex';

let GlobalStore = useStore();
let ActiveClass = computed(() => {
	if (GlobalStore.getters.GetScreenLevel < 2)
		return 'topMenu'
	else
		return 'leftMenu'
})
</script>
