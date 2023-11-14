<template>
	<div id="sliding" :class="ActiveClass" class="duration-300 flex justify-between">
		<div id="menu" class="py-8 pl-8">
			<h1> Ana Sayfa </h1>
			<h1> Menu </h1>
		</div>
		<div id="iconBar" :class="IconBarClass" class="flex justify-center items-center">
			<v-icon name="hi-menu" class="menuIcon"></v-icon>
		</div>
	</div>
</template>

<style scoped>
#menu{
	width: 95%;
	height: 95%;
}
</style>

<script setup lang="ts">
import { computed, inject } from 'vue';
import { useStore } from 'vuex';
import type { ServiceType } from '@/service/service'

let service:ServiceType = inject<ServiceType>('Service');
let a = async () => {
	console.log('inject : ', service && await service.getMenu());
}
a();
let GlobalStore = useStore();
let ActiveClass = computed(() => { return GlobalStore.getters.GetScreenLevel < 2 ? 'topMenu flex-col' : 'leftMenu flex-row'; })
let IconBarClass = computed(() => { return ActiveClass.value == 'topMenu flex-col' ? 'topIcon' : 'leftIcon'; })
</script>
