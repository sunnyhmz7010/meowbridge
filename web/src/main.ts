import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { themeStore } from './stores/theme'
import './style.css'

themeStore.init()

createApp(App).use(router).mount('#app')
