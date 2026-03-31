<script setup lang="ts">
import { ref, onMounted, provide } from 'vue'
import Layout from './components/Layout.vue'
import { apiClient, wsClient } from './api/client'

// Global state
const systemInfo = ref<any>(null)
const agents = ref<any[]>([])
const tasks = ref<any[]>([])
const isConnected = ref(false)

// Provide to all components
provide('systemInfo', systemInfo)
provide('agents', agents)
provide('tasks', tasks)
provide('isConnected', isConnected)

onMounted(async () => {
  try {
    systemInfo.value = await apiClient.getSystemInfo()
    agents.value = (await apiClient.getAgents()).agents || []
    tasks.value = (await apiClient.getTasks()).tasks || []
  } catch (e) {
    console.error('Failed to load initial data:', e)
  }
})
</script>

<template>
  <Layout>
    <router-view />
  </Layout>
</template>