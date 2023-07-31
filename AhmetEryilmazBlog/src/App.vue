<template>
	<teleport to='#sidemenu'>
		<sidemenu></sidemenu>
	</teleport>
	<div id="content-grid" class="grid grid-cols-10 h-screen">
		<div class="col-span-7"></div>
		<contentView.viewProfile class="col-span-3" />
	</div>
</template>

<style scoped>
#content-grid {
	margin-left: 2%;
}
</style>

<script setup lang="ts">
import sidemenu from '@/views/SideMenu.vue'
import * as contentView from '@/components/contentViews'
</script>
