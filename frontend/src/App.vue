<template>
	<teleport to='#sidemenu'>
		<suspense>
			<routes.sideMenu
			/>
		</suspense>
	</teleport>
	<div id="content-grid" class="flex h-screen" :class="GlobalStore.getters.GetIsMobile ? 'flex-col mt-8' : 'flex-row  ml-8'">
		<routes.sidePanel v-if="GlobalStore.getters.GetIsMobile" class="bg-surface border-2 border-border rounded mx-4 mt-4 mb-2 basis-1/4 h-auto"/>
		<!-- <routes.contents -->
		<RouterView
			class="bg-surface border-2 border-border rounded"
			:class="GlobalStore.getters.GetIsMobile ? 'basis-full mx-4 mt-2 mb-4' : 'basis-3/4 m-4'"
		>
		</RouterView>
		<routes.sidePanel v-if="!GlobalStore.getters.GetIsMobile" class="bg-surface border-2 border-border rounded m-4 basis-1/4"/>
	</div>
</template>

<style scoped>
</style>

<script setup lang="ts">
import * as routes from '@/router/routes'
import { useStore } from 'vuex';
import { onMounted, provide } from 'vue'
import serviceClass from '@/service/BaseAPI.service'
provide('Service', serviceClass);

let GlobalStore = useStore();

let handleResize = function() {
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth)
}
onMounted(() => {
	window.addEventListener('resize', handleResize);
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth);
	GlobalStore.dispatch('SetActivePage', "MainPage");
})
</script>
@/service/BaseAPI.service
