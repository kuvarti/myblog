<template>
	<div class="py-4 pl-4 pr-4">
		<div class="content" v-html="returnedHTML"></div>
	</div>
</template>

<style scoped>
* {
	overflow-y: auto;
}
</style>

<script setup lang="ts">
import { onMounted, ref, inject, watch } from 'vue';
import { useStore } from 'vuex';
import { type ServiceType } from '@/service/BaseAPI.service'

let service:ServiceType = inject<ServiceType>('Service');
let GlobalStore = useStore()
let returnedHTML = ref<string>("");

function fetchPage(name: string) {
	service?.getPage(name).then((data) => {
		returnedHTML.value = data.Page;
	}).catch((err) => {
		console.error(err);
		returnedHTML.value = "Error";
	})
}

onMounted(() => fetchPage(GlobalStore.getters.GetActivePage))
watch(() => GlobalStore.getters.GetActivePage, (name) => fetchPage(name))
</script>
@/service/BaseAPI.service
