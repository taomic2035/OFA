<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { apiClient } from '../api/client'
import type { SystemMetrics } from '../types'

const metrics = ref<SystemMetrics | null>(null)
const loading = ref(true)
const autoRefresh = ref(true)
let refreshInterval: ReturnType<typeof setInterval> | null = null

async function loadMetrics() {
  try {
    metrics.value = await apiClient.getMetrics()
  } catch (e) {
    console.error('Failed to load metrics:', e)
  } finally {
    loading.value = false
  }
}

function toggleAutoRefresh() {
  autoRefresh.value = !autoRefresh.value
  if (autoRefresh.value) {
    refreshInterval = setInterval(loadMetrics, 5000)
  } else if (refreshInterval) {
    clearInterval(refreshInterval)
    refreshInterval = null
  }
}

onMounted(() => {
  loadMetrics()
  if (autoRefresh.value) {
    refreshInterval = setInterval(loadMetrics, 5000)
  }
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2>Monitor</h2>
      <div class="flex gap-2">
        <button class="btn btn-outline" @click="toggleAutoRefresh">
          <span :class="['badge', autoRefresh ? 'badge-success' : 'badge-gray']" style="margin-right: 0.5rem;"></span>
          Auto Refresh
        </button>
        <button class="btn btn-outline" @click="loadMetrics">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
          Refresh
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center p-8">
      <p>Loading metrics...</p>
    </div>

    <div v-else-if="metrics">
      <!-- Agent Metrics -->
      <div class="card mb-4">
        <div class="card-header">
          <h3 class="card-title">Agent Metrics</h3>
        </div>
        <div class="stats-grid" style="margin-bottom: 0;">
          <div class="stat-card">
            <div class="label">Total Agents</div>
            <div class="value">{{ metrics.agents?.total || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Online</div>
            <div class="value text-success">{{ metrics.agents?.online || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Offline</div>
            <div class="value text-danger">{{ metrics.agents?.offline || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Busy</div>
            <div class="value text-warning">{{ metrics.agents?.busy || 0 }}</div>
          </div>
        </div>
      </div>

      <!-- Task Metrics -->
      <div class="card mb-4">
        <div class="card-header">
          <h3 class="card-title">Task Metrics</h3>
        </div>
        <div class="stats-grid" style="margin-bottom: 0;">
          <div class="stat-card">
            <div class="label">Total Tasks</div>
            <div class="value">{{ metrics.tasks?.total || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Pending</div>
            <div class="value">{{ metrics.tasks?.pending || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Running</div>
            <div class="value text-warning">{{ metrics.tasks?.running || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Completed</div>
            <div class="value text-success">{{ metrics.tasks?.completed || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Failed</div>
            <div class="value text-danger">{{ metrics.tasks?.failed || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Success Rate</div>
            <div class="value">
              {{ metrics.tasks?.total ? Math.round((metrics.tasks.completed / metrics.tasks.total) * 100) : 0 }}%
            </div>
          </div>
        </div>
      </div>

      <!-- Message Metrics -->
      <div class="card mb-4">
        <div class="card-header">
          <h3 class="card-title">Message Metrics</h3>
        </div>
        <div class="stats-grid" style="margin-bottom: 0;">
          <div class="stat-card">
            <div class="label">Total Messages</div>
            <div class="value">{{ metrics.messages?.total || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Delivered</div>
            <div class="value text-success">{{ metrics.messages?.delivered || 0 }}</div>
          </div>
          <div class="stat-card">
            <div class="label">Pending</div>
            <div class="value">{{ metrics.messages?.pending || 0 }}</div>
          </div>
        </div>
      </div>

      <!-- Performance Metrics -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">Performance</h3>
        </div>
        <div class="stats-grid" style="margin-bottom: 0;">
          <div class="stat-card">
            <div class="label">Avg Task Duration</div>
            <div class="value">{{ Math.round(metrics.performance?.avgTaskDuration || 0) }}ms</div>
          </div>
          <div class="stat-card">
            <div class="label">Avg Response Time</div>
            <div class="value">{{ Math.round(metrics.performance?.avgResponseTime || 0) }}ms</div>
          </div>
          <div class="stat-card">
            <div class="label">Throughput</div>
            <div class="value">{{ Math.round(metrics.performance?.throughput || 0) }}/s</div>
          </div>
        </div>
      </div>
    </div>

    <div v-else class="empty-state">
      <p>Failed to load metrics</p>
      <button class="btn btn-primary mt-4" @click="loadMetrics">Retry</button>
    </div>
  </div>
</template>