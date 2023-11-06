<template>
	<teleport to='#sidemenu'>
		<routes.sideMenu />
	</teleport>
	<div id="content-grid" class="">
		<routes.contents class="" />
		<routes.viewProfile v-if="GlobalStore.getters.GetScreenLevel > 2" class="" />
	</div>
</template>

<style scoped>
#content-grid {
	margin-left: 2%;
}
</style>

<script setup lang="ts">
import * as routes from '@/router/routes'
import { useStore } from 'vuex';
import { onMounted, ref } from 'vue';

let GlobalStore = useStore();
let handleResize = function() {
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth)
}

onMounted: {
	window.addEventListener('resize', handleResize);
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth)
}

</script>
