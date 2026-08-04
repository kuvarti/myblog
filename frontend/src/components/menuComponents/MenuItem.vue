<template>
	<!-- <router-link :to="RouterRedirect"> -->
		<div
			:class="textColorFunction"
			class="text-4xl my-2 ml-4 hover:ml-6 hover:cursor-pointer hover:text-accent"
			@click="RouterRedirect"
		>
			{{ Caption }}
		</div>
	<!-- </router-link> -->
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { computed } from 'vue'
import { type MenuListModal } from '@/types/MenuListModal'
import { isMenuItemActive } from '@/components/menuComponents/menuActive'
import router from '@/router';

let route = useRoute();
let props = defineProps<MenuListModal>()
let textColorFunction = computed(() =>
	isMenuItemActive(props, route.path)
		? 'text-activePageColor'
		: 'text-deActivePageColor'
)

// Per-page routing: navigate straight to the page's own Path. A Path-less item
// falls back to the home route.
let RouterRedirect = () => {
	router.push(props.Path || '/')
}
</script>
