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
import { useRoute } from 'vue-router';
import { type ServiceType } from '@/service/BaseAPI.service'

let service:ServiceType = inject<ServiceType>('Service');
let route = useRoute();
let returnedHTML = ref<string>("");

function fetchPage(path: string) {
	service?.getPageByPath(path).then((data) => {
		returnedHTML.value = data.Page;
	}).catch((err) => {
		console.error(err);
		returnedHTML.value = "Error";
	})
}

onMounted(() => fetchPage(route.path))
watch(() => route.path, (path) => fetchPage(path))
</script>
