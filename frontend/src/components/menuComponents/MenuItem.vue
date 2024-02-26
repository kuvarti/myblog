<template>
	<router-link :to="Path || '/'">
		<div
			:class="textColorFunction"
			class="text-4xl my-2 ml-4 hover:ml-6 hover:cursor-pointer hover:text-activePageColor"
		>
			{{ Caption }}
		</div>
	</router-link>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import { type MenuListModal } from '@/types/MenuListModal'

let route = useRoute();
let props = defineProps<MenuListModal>()
let currentRoutePath = computed(() => route.path);
let textColorFunction = computed(() => {
	let propPath: string[] = props.Path?.split('/') || [];
	let currentPath: string[] = currentRoutePath.value.split('/');

	if (currentPath[1] == propPath[1])
		return 'text-activePageColor'
	else
		return 'text-deActivePageColor'
})
</script>
