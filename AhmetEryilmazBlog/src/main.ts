import './assets/main.css'
import './assets/sliderMenu.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import vuex from '@/global/store'
import icons from '@/global/vueIcons'


const app = createApp(App)
app.use(router).use(vuex)
app.component('v-icon', icons);
app.mount('#app');
