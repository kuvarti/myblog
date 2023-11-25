<template>
	<div id="sliding" :class="ActiveClass" class="duration-300 flex justify-between">
		<div id="menu" class="py-8 pl-8 w-3/4 h-3/4">
			<div v-if="MenuList.length > 0"  class="flex flex-col">
				<div v-for="items in MenuList" :key="items.Name">
					<MenuItem :Path="items.Path" :Name="items.Name" :Caption="items.Caption"/>
				</div>
			</div>
			<div v-else>
				<MenuItem Path="/" Name="contents" Caption="Ana Sayfa" />
			</div>
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

<script async setup lang="ts">
import { computed, inject, ref, onMounted, watchEffect } from 'vue';
import { useStore } from 'vuex';
import { type ServiceType } from '@/service/service'
import { type MenuListModal } from '@/types/MenuListModal'
import MenuItem from '@/components/menuComponents/MenuItem.vue'

let service:ServiceType = inject<ServiceType>('Service');

let GlobalStore = useStore();
let ActiveClass = computed(() => { return GlobalStore.getters.GetScreenLevel < 2 ? 'topMenu flex-col' : 'leftMenu flex-row'; })
let IconBarClass = computed(() => { return ActiveClass.value == 'topMenu flex-col' ? 'topIcon' : 'leftIcon'; })
let MenuList = ref<MenuListModal[]>([]);

async function getData() {
	if (service && typeof service.getMenu === 'function') {
		return await service.getMenu()
	} else {
		console.error('Service or getMenu method not available');
		return [];
	}
}

onMounted(() => {
	watchEffect(async () => {
		MenuList.value = await service?.getMenu() || [];
	});
});
</script>
