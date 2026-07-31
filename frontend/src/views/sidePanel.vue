<template>
	<div id="profile-grid" class="" v-if="!IsPanel"><!--overflow-hidden-->
		<div class="flex justify-end p-3">
			<ThemeToggle />
		</div>
		<contentView.viewProfileUpper class="px-4 pb-4" v-if="GlobalStore.getters.GetIsMobile"/>
		<div class="" v-if="!GlobalStore.getters.GetIsMobile">
			<contentView.viewProfileUpper class="p-4"/>
			<hr>
			<contentView.lastActivity/>
		</div>
	</div>
	<div v-else class="flex-nowrap">
		<Menu class=""/>
	</div>
</template>

<style scoped>
hr {
	border: none;
	border-top: 1px solid var(--border);
	margin:  0 2.5em 0 2.5em;
	position: relative;
}
</style>

<script setup lang="ts">
import * as contentView from '@/components/contentViews'
import { Menu } from '@/components/panelViews'
import ThemeToggle from '@/components/ThemeToggle.vue'
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useStore } from 'vuex';

let GlobalStore = useStore();
let IsPanel = computed(() => {
	let deneme = useRoute();
	return deneme.fullPath.includes("panel");
})
</script>
