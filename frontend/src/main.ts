import './assets/main.css'
import './assets/sliderMenu.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import vuex from '@/global/store'
import icons from '@/global/vueIcons'
import PrimeVue from 'primevue/config';
import 'primeicons/primeicons.css'


const app = createApp(App)
app.use(router).use(vuex).use(PrimeVue)
app.component('v-icon', icons);
app.mount('#app');
