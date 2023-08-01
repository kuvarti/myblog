<template>
	<teleport to='#sidemenu'>
		<routes.sideMenu />
	</teleport>
	<div id="content-grid" class="grid grid-cols-10 grid-flow-row h-screen">
		<routes.contents class="col-span-7" />
		<routes.viewProfile class="col-span-3" />
	</div>
</template>

<style scoped>
#content-grid {
	margin-left: 2%;
}
</style>

<script setup lang="ts">
import * as routes from '@/router/routes'
</script>
