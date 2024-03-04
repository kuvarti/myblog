<template>
	<div class="flex flex-row justify-around h-full w-full">
		<PanelMenu :model="items" class="text-white overflow-auto w-full" v-if="IsAuth">
			<template #item="item">
				<a class="flex items-center px-3 py-2 cursor-pointer border hover:bg-slate-900/50">
					<span :class="[item.item.icon, 'text-primary-500 dark:text-primary-400']" />
					<span :class="['ml-2', { 'font-semibold hover:bg-black': item.item.items }]">{{ item.label }}</span>
				</a>
			</template>
		</PanelMenu>
	</div>
</template>

<style scoped>
.flex a {
  padding: 10px 15px; /* Artırılmış iç boşluk */
  font-size: 16px; /* Büyük yazı tipi boyutu */
  height: /*75px*/100px; /* Artırılmış yükseklik */
}

.flex span {
  font-size: 20px; /* İkon ve metin için büyük yazı tipi boyutu */
}

@media (max-width <= 1280) {
	.flex a {
		padding: 10px 15px; /* Artırılmış iç boşluk */
		font-size: 16px; /* Büyük yazı tipi boyutu */
		height: /*75px*/75px; /* Artırılmış yükseklik */
	}
}
</style>

<script setup lang="ts">
import 'primeicons/primeicons.css'
import { ref, provide } from "vue"
import PanelMenu from 'primevue/panelmenu';
import UserService, {type UserServiceType} from '@/service/User.service';
import { useStore } from 'vuex';
provide<UserServiceType>("UserService", UserService)

let GlobalStore = useStore();
let IsAuth = UserService.IsLogin;
const items = [
	{
		label: 'File',
		icon: 'pi pi-fw pi-file'
	},
	{
		label: 'Edit',
		icon: 'pi pi-fw pi-pencil'
	},
	{
		label: 'File',
		icon: 'pi pi-fw pi-file'
	},
	{
		label: 'Edit',
		icon: 'pi pi-fw pi-pencil'
	},
	{
		label: 'File',
		icon: 'pi pi-fw pi-file'
	},
	{
		label: 'Edit',
		icon: 'pi pi-fw pi-pencil'
	},
];
</script>
