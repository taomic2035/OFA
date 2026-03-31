import { createApp } from 'vue'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import './styles/main.css'

// Views
import Dashboard from './views/Dashboard.vue'
import Agents from './views/Agents.vue'
import Tasks from './views/Tasks.vue'
import Monitor from './views/Monitor.vue'
import Messages from './views/Messages.vue'

const routes = [
  { path: '/', name: 'Dashboard', component: Dashboard },
  { path: '/agents', name: 'Agents', component: Agents },
  { path: '/tasks', name: 'Tasks', component: Tasks },
  { path: '/monitor', name: 'Monitor', component: Monitor },
  { path: '/messages', name: 'Messages', component: Messages }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const app = createApp(App)
app.use(router)
app.mount('#app')