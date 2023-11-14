<template>
	<div id="profile-grid" class=""><!--overflow-hidden-->
		<contentView.viewProfileUpper class="" v-if="GlobalStore.getters.GetIsMobile"/>
		<div class="" v-if="!GlobalStore.getters.GetIsMobile">
			<contentView.viewProfileUpper class="p-4"/>
			<hr>
			<contentView.lastActivity/>
		</div>
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
import { useStore } from 'vuex';

let GlobalStore = useStore();
</script>
