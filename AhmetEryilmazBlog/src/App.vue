<template>
	<teleport to='#sidemenu'>
		<suspense>
			<routes.sideMenu
				@ContentComponent="changeComponent($event)"
			/>
		</suspense>
	</teleport>
	<div id="content-grid" class="flex h-screen" :class="GlobalStore.getters.GetIsMobile ? 'flex-col mt-8' : 'flex-row  ml-8'">
		<routes.viewProfile v-if="GlobalStore.getters.GetIsMobile" class="bg-midnightPurple border-2 rounded m-4 basis-1/4" :class="GlobalStore.getters.GetIsMobile ? 'h-8' : 'h-screen'"/>
		<!-- <routes.contents -->
		<component
			:is="getComponent"
			class="bg-midnightPurple border-2 rounded m-4"
			:class="GlobalStore.getters.GetIsMobile ? 'basis-full' : 'basis-3/4'"
		/>
		<routes.viewProfile v-if="!GlobalStore.getters.GetIsMobile" class="bg-midnightPurple border-2 rounded m-4 basis-1/4"/>
	</div>
</template>

<style scoped>
</style>

<script setup lang="ts">
import * as routes from '@/router/routes'
import pageArray from '@/router/routes'
import { useStore } from 'vuex';
import { computed, onMounted, provide, shallowRef, watch } from 'vue'
import serviceClass from '@/service/service'
provide('Service', serviceClass);

let GlobalStore = useStore();
let getComponent = shallowRef();

let handleResize = function() {
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth)
}
function setComponent(page: string) {
	let ret = pageArray.find(elem => elem.__name == page);
	getComponent.value = ret || getComponent.value;
	return ret;
}

onMounted(() => {
	window.addEventListener('resize', handleResize);
	GlobalStore.dispatch('SetScreenLevel', window.innerWidth);
	GlobalStore.dispatch('SetActivePage', "mainPage");
	setComponent('contents');
})

function changeComponent(data: any) {
	setComponent(data.name);
}
</script>
