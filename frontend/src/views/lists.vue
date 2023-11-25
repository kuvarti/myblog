<template>
	<div class="py-4 pl-4">
		<div class="">
			<p> Lorem ipsum dolor sit amet consectetur adipisicing elit. Commodi ab optio quae voluptatum aut, tenetur dolores. Totam, non reiciendis. Sed ab aperiam natus numquam totam eum incidunt consectetur obcaecati itaque! </p>
		</div>
	</div>
</template>

<style scoped>
* {
	overflow-y: auto;
}
</style>

<script setup lang="ts">
import { computed } from 'vue';
import { useStore } from 'vuex';

let GlobalStore = useStore()
const returnedHTML = computed(() => {
	return GlobalStore.getters.IsMobile
})
</script>
