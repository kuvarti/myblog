import './assets/main.css'
import './assets/sliderMenu.css'

import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import vuex from '@/global/store'

const app = createApp(App)

app.use(router).use(vuex)

app.mount('#app')
