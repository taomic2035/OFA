<script setup lang="ts">
import { ref, onMounted, inject } from 'vue'
import { apiClient } from '../api/client'
import type { SystemInfo, SystemMetrics } from '../types'

const systemInfo = inject<any>('systemInfo')
const agents = inject<any>('agents')
const tasks = inject<any>('tasks')

const metrics = ref<SystemMetrics | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    metrics.value = await apiClient.getMetrics()
  } catch (e) {
    console.error('Failed to load metrics:', e)
  } finally {
    loading.value = false
  }
})

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}d ${hours}h ${mins}m`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'online': return 'badge-success'
    case 'offline': return 'badge-danger'
    case 'busy': return 'badge-warning'
    default: return 'badge-gray'
  }
}
</script>

<template>
  <div>
    <h2 style="margin-bottom: 1.5rem;">Dashboard</h2>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="label">Total Agents</div>
        <div class="value">{{ metrics?.agents?.total || 0 }}</div>
        <div class="change positive">{{ metrics?.agents?.online || 0 }} online</div>
      </div>
      <div class="stat-card">
        <div class="label">Total Tasks</div>
        <div class="value">{{ metrics?.tasks?.total || 0 }}</div>
        <div class="change">{{ metrics?.tasks?.completed || 0 }} completed</div>
      </div>
      <div class="stat-card">
        <div class="label">Running Tasks</div>
        <div class="value">{{ metrics?.tasks?.running || 0 }}</div>
        <div class="change">{{ metrics?.tasks?.pending || 0 }} pending</div>
      </div>
      <div class="stat-card">
        <div class="label">Failed Tasks</div>
        <div class="value text-danger">{{ metrics?.tasks?.failed || 0 }}</div>
        <div class="change negative">{{ metrics?.tasks?.failed ? 'Needs attention' : 'All good' }}</div>
      </div>
    </div>

    <!-- Agent Status -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">Recent Agents</h3>
        <router-link to="/agents" class="btn btn-outline btn-sm">View All</router-link>
      </div>
      <div class="table-container">
        <table v-if="agents?.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Type</th>
              <th>Status</th>
              <th>Load</th>
              <th>Last Seen</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="agent in agents?.slice(0, 5)" :key="agent.id">
              <td><code>{{ agent.id?.slice(0, 8) }}</code></td>
              <td>{{ agent.name }}</td>
              <td>{{ agent.type }}</td>
              <td><span :class="['badge', getStatusClass(agent.status)]">{{ agent.status }}</span></td>
              <td>{{ agent.load }}%</td>
              <td class="text-muted">{{ new Date(agent.lastSeen).toLocaleString() }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">
          <p>No agents connected</p>
        </div>
      </div>
    </div>

    <!-- Recent Tasks -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">Recent Tasks</h3>
        <router-link to="/tasks" class="btn btn-outline btn-sm">View All</router-link>
      </div>
      <div class="table-container">
        <table v-if="tasks?.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>Skill</th>
              <th>Status</th>
              <th>Agent</th>
              <th>Created</th>
              <th>Duration</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in tasks?.slice(0, 5)" :key="task.id">
              <td><code>{{ task.id?.slice(0, 8) }}</code></td>
              <td>{{ task.skillId }}</td>
              <td><span :class="['badge', getStatusClass(task.status)]">{{ task.status }}</span></td>
              <td>{{ task.agentId?.slice(0, 8) || '-' }}</td>
              <td class="text-muted">{{ new Date(task.createdAt).toLocaleString() }}</td>
              <td>{{ task.duration ? `${task.duration}ms` : '-' }}</td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">
          <p>No tasks yet</p>
        </div>
      </div>
    </div>

    <!-- System Info -->
    <div class="card" v-if="systemInfo">
      <div class="card-header">
        <h3 class="card-title">System Information</h3>
      </div>
      <div class="grid grid-4">
        <div>
          <div class="text-muted mb-1">Version</div>
          <div><strong>{{ systemInfo.version }}</strong></div>
        </div>
        <div>
          <div class="text-muted mb-1">Uptime</div>
          <div><strong>{{ formatUptime(systemInfo.uptime) }}</strong></div>
        </div>
        <div>
          <div class="text-muted mb-1">Agents Online</div>
          <div><strong>{{ systemInfo.agentsOnline }} / {{ systemInfo.agentsTotal }}</strong></div>
        </div>
        <div>
          <div class="text-muted mb-1">Tasks Completed</div>
          <div><strong>{{ systemInfo.tasksCompleted }} / {{ systemInfo.tasksTotal }}</strong></div>
        </div>
      </div>
    </div>
  </div>
</template>