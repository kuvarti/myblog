<template>
	<div id="profile-grid" class="p-16">
		<div class="rounded border-2 border-white h-full w-full grid grid-rows-4">
			<contentView.viewProfileUpper class="h-48 p-4" />
			<div class="h-full w-full row-span-3"></div>
		</div>
	</div>
</template>

<style scoped></style>

<script setup lang="ts">
import * as contentView from '@/components/contentViews'
</script>