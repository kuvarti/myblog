<template>
	<div v-if="!AuthChecked" class="p-6 text-muted">Loading…</div>
	<div v-else-if="!IsAuth">
		<panelView.Login />
	</div>
	<div v-else class="h-full">
		<PageEditor />
	</div>
</template>

<script async setup lang="ts">
import * as panelView from "@/components/panelViews"
import PageEditor from "@/components/panelComponents/PageEditor.vue"
import UserService, { type UserServiceType } from "@/service/User.service"
import { provide } from "vue"
provide<UserServiceType>("UserService", UserService)

const IsAuth = UserService.IsLogin
const AuthChecked = UserService.AuthChecked
</script>
