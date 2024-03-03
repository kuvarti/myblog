<template>
	<div v-if="!IsAuth">
		<panelView.Login />
	</div>
	<div v-else>
		{{ IsAuth }}
	</div>
</template>

<script async setup lang="ts">
import * as panelView from "@/components/panelViews"
import UserService, { type UserServiceType } from "@/service/User.service"
import { computed, provide } from "vue";
provide<UserServiceType>("UserService", UserService)

let IsAuth = UserService.IsLogin;
</script>
