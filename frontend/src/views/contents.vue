<template>
	<div class="py-4 pl-4 pr-4">
		<div class="content" v-html="returnedHTML" @click="onContentClick"></div>
	</div>
</template>

<style scoped>
* {
	overflow-y: auto;
}
</style>

<script setup lang="ts">
import { onMounted, ref, inject, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { type ServiceType } from '@/service/BaseAPI.service'
import { interceptNavTarget } from '@/components/contentLinks'

let service:ServiceType = inject<ServiceType>('Service');
let route = useRoute();
const router = useRouter();
let returnedHTML = ref<string>("");

function fetchPage(path: string) {
	service?.getPageByPath(path).then((data) => {
		returnedHTML.value = data.Page;
	}).catch((err) => {
		console.error(err);
		returnedHTML.value = "Error";
	})
}

function onContentClick(e: MouseEvent) {
	const el = e.target as HTMLElement
	const a = el.closest('a')
	if (!a) return
	const to = interceptNavTarget(e, a)
	if (to) {
		e.preventDefault()
		router.push(to)
	}
}

onMounted(() => fetchPage(route.path))
watch(() => route.path, (path) => fetchPage(path))
</script>
