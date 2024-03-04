<template>
	<div id="profile-grid" class="" v-if="!IsPanel"><!--overflow-hidden-->
		<contentView.viewProfileUpper class="" v-if="GlobalStore.getters.GetIsMobile"/>
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
	border-top: 1px solid rgba(255,255,255,.3);
	border-bottom: 1px solid rgba(0,0,0,.08);
	margin:  0 2.5em 0 2.5em;
	position: relative;
}
</style>

<script setup lang="ts">
import * as contentView from '@/components/contentViews'
import { Menu } from '@/components/panelViews'
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useStore } from 'vuex';

let GlobalStore = useStore();
let IsPanel = computed(() => {
	let deneme = useRoute();
	return deneme.fullPath.includes("panel");
})
</script>
