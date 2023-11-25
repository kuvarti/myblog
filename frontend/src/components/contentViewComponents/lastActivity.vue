<template>
	<!-- <div v-for="items in firstThreeItems" :key="items.ID" :content="items">
		{{ items.value }}
	</div> -->
</template>

<script setup lang="ts">
import { useStore } from 'vuex';

let GlobalStore = useStore();
let lastActivityArray = [
	{
		ID: 1,
		value: "text1"
	},
	{
		ID: 2,
		value: "text2"
	},
	{
		ID: 3,
		value: "text3"
	},
	{
		ID: 4,
		value: "text4"
	}
];
const firstThreeItems = lastActivityArray.slice(0, 3);
</script>
