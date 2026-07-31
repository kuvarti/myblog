<template>
	<div class="flex flex-col items-center justify-center px-6 py-8 mx-auto md:h-screen lg:py-0">
		<div class="w-full bg-surface border border-border rounded-lg shadow md:mt-0 sm:max-w-md xl:p-0">
			<div class="p-6 space-y-4 md:space-y-6 sm:p-8">
				<form class="space-y-4 md:space-y-6" @submit.prevent="SignIn">
					<div>
						<label for="user" class="block mb-2 text-sm font-medium text-fg">Username</label>
						<InputText type="text" v-model="form.user" class="bg-surface-2 border border-border text-fg sm:text-sm rounded-lg focus:ring-accent focus:border-accent block w-full p-2.5" placeholder="UserName" required/>
					</div>
					<div>
						<label for="password" class="block mb-2 text-sm font-medium text-fg">Password</label>
						<InputText type="password" v-model="form.pass" class="bg-surface-2 border border-border text-fg sm:text-sm rounded-lg focus:ring-accent focus:border-accent block w-full p-2.5" placeholder="PassWord" required/>
					</div>
					<Button label="Log In" severity="contrast" type="submit" class="w-full bg-accent text-surface border border-border rounded hover:opacity-90"/>
				</form>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import type { UserServiceType } from '@/service/User.service';
import Password from 'primevue/password';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
import { inject, ref } from 'vue';

let form = ref({
	user: "",
	pass: ""
})
let UserService = inject<UserServiceType>("UserService")
let SignIn = async () => {
	UserService?.Login(form.value.user, form.value.pass)
}
</script>
