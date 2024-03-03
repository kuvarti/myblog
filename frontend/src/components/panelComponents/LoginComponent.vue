<template>
	<div class="flex flex-col items-center justify-center px-6 py-8 mx-auto md:h-screen lg:py-0">
		<div class="w-full bg-gray-500 rounded-lg shadow dark:border md:mt-0 sm:max-w-md xl:p-0 dark:bg-gray-800 dark:border-gray-700">
			<div class="p-6 space-y-4 md:space-y-6 sm:p-8">
				<form class="space-y-4 md:space-y-6" @submit.prevent="SignIn">
					<div>
						<label for="user" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Username</label>
						<InputText type="text" v-model="form.user" class="bg-gray-50 border border-gray-300 text-gray-900 sm:text-sm rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="UserName" required/>
					</div>
					<div>
						<label for="password" class="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Password</label>
						<InputText type="password" v-model="form.pass" class="bg-gray-50 border border-gray-300 text-gray-900 sm:text-sm rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" placeholder="PassWord" required/>
					</div>
					<Button label="Log In" severity="contrast" type="submit" class="w-full bg-gray-300 border border-gray-400 rounded hover:bg-gray-400"/>
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
