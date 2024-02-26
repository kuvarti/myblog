<template>
	<div class="py-4 pl-4 pr-4">
		<div class="">
			<div v-html="returnedHTML"></div>
		</div>
	</div>
</template>

<style scoped>
* {
	overflow-y: auto;
}
</style>

<script setup lang="ts">
import { computed, onMounted, ref, inject } from 'vue';
import { useStore } from 'vuex';
import { type ServiceType } from '@/service/service'

let service:ServiceType = inject<ServiceType>('Service');
let GlobalStore = useStore()
let returnedHTML = ref<string>("");

onMounted(() => {
	service?.getPage(GlobalStore.getters.GetActivePage).then((data) => {
		returnedHTML.value = data.Page;
	}).catch((err) => {
		console.error(err);
		returnedHTML.value = "Error";
	})
})
</script>
