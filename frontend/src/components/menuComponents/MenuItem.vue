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
import { useStore } from 'vuex'
import { type MenuListModal } from '@/types/MenuListModal'
import { isMenuItemActive } from '@/components/menuComponents/menuActive'
import router from '@/router';

let route = useRoute();
let GlobalStore = useStore();
let props = defineProps<MenuListModal>()
let activePage = computed(() => GlobalStore.getters.GetActivePage as string);
let textColorFunction = computed(() =>
	isMenuItemActive(props, route.path, activePage.value)
		? 'text-activePageColor'
		: 'text-deActivePageColor'
)

let RouterRedirect = () => {
	if (props.PageName) {
		GlobalStore.dispatch('SetActivePage', props.PageName)
		router.push('/')
	} else {
		router.push(props.Path || '/')
	}
}
</script>
