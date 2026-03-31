<script setup lang="ts">
import { ref, onMounted, inject, onUnmounted } from 'vue'
import { apiClient, wsClient } from '../api/client'
import type { SystemInfo, SystemMetrics, Agent, Task } from '../types'

const systemInfo = ref<SystemInfo | null>(null)
const metrics = ref<SystemMetrics | null>(null)
const agents = inject<any>('agents')
const tasks = inject<any>('tasks')
const loading = ref(true)
const isConnected = ref(false)

// Refresh data periodically
let refreshInterval: ReturnType<typeof setInterval> | null = null

async function loadAllData() {
  try {
    const [info, metricsData, agentsData, tasksData] = await Promise.all([
      apiClient.getSystemInfo(),
      apiClient.getMetrics(),
      apiClient.getAgents(),
      apiClient.getTasks()
    ])
    systemInfo.value = info
    metrics.value = metricsData
    agents.value = agentsData.agents || []
    tasks.value = tasksData.tasks || []
  } catch (e) {
    console.error('Failed to load data:', e)
  } finally {
    loading.value = false
  }
}

function formatUptime(seconds: number): string {
  if (!seconds) return '0m'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'online':
    case 'completed': return 'badge-success'
    case 'offline':
    case 'failed':
    case 'cancelled': return 'badge-danger'
    case 'busy':
    case 'running': return 'badge-warning'
    default: return 'badge-info'
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleTimeString()
}

onMounted(async () => {
  await loadAllData()

  // Auto refresh every 10 seconds
  refreshInterval = setInterval(loadAllData, 10000)

  // WebSocket connection
  wsClient.connect()
  const unsubConnect = wsClient.on('connected', () => {
    isConnected.value = true
  })
  const unsubDisconnect = wsClient.on('disconnected', () => {
    isConnected.value = false
  })
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
  wsClient.disconnect()
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold" style="margin: 0;">控制台</h1>
        <p class="text-muted mt-1" style="margin: 0;">OFA 系统概览</p>
      </div>
      <div class="flex items-center gap-4">
        <div class="flex items-center gap-2 text-sm">
          <span :class="['status-dot', isConnected ? 'connected' : 'disconnected']"></span>
          {{ isConnected ? '已连接' : '未连接' }}
        </div>
        <button class="btn btn-outline" @click="loadAllData" :disabled="loading">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
          刷新
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- Content -->
    <template v-else>
      <!-- Stats Cards -->
      <div class="stats-grid">
        <div class="stat-card gradient-blue">
          <div class="stat-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
              <path d="M23 21v-2a4 4 0 0 0-3-3.87"/>
              <path d="M16 3.13a4 4 0 0 1 0 7.75"/>
            </svg>
          </div>
          <div class="stat-content">
            <div class="stat-label">智能体</div>
            <div class="stat-value">{{ metrics?.agents?.total || 0 }}</div>
            <div class="stat-detail">
              <span class="text-success">{{ metrics?.agents?.online || 0 }} 在线</span>
              <span class="text-muted ml-2">/ {{ metrics?.agents?.offline || 0 }} 离线</span>
            </div>
          </div>
        </div>

        <div class="stat-card gradient-green">
          <div class="stat-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
            </svg>
          </div>
          <div class="stat-content">
            <div class="stat-label">任务</div>
            <div class="stat-value">{{ metrics?.tasks?.total || 0 }}</div>
            <div class="stat-detail">
              <span class="text-success">{{ metrics?.tasks?.completed || 0 }} 完成</span>
              <span class="text-warning ml-2">{{ metrics?.tasks?.running || 0 }} 运行</span>
            </div>
          </div>
        </div>

        <div class="stat-card gradient-purple">
          <div class="stat-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <div class="stat-content">
            <div class="stat-label">消息</div>
            <div class="stat-value">{{ metrics?.messages?.total || 0 }}</div>
            <div class="stat-detail">
              <span class="text-success">{{ metrics?.messages?.delivered || 0 }} 已送达</span>
            </div>
          </div>
        </div>

        <div class="stat-card gradient-orange">
          <div class="stat-icon">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <polyline points="12 6 12 12 16 14"/>
            </svg>
          </div>
          <div class="stat-content">
            <div class="stat-label">运行时间</div>
            <div class="stat-value text-lg">{{ formatUptime(systemInfo?.uptime || 0) }}</div>
            <div class="stat-detail">
              <span class="text-muted">v{{ systemInfo?.version || '0.0.0' }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Two Column Layout -->
      <div class="dashboard-grid">
        <!-- Recent Agents -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
                <circle cx="9" cy="7" r="4"/>
              </svg>
              最近智能体
            </h3>
            <router-link to="/agents" class="btn btn-outline btn-sm">查看全部</router-link>
          </div>
          <div class="card-body">
            <div v-if="agents?.length" class="agent-list">
              <div v-for="agent in agents.slice(0, 5)" :key="agent.id" class="agent-item">
                <div class="agent-avatar" :class="agent.status">
                  {{ agent.name?.charAt(0)?.toUpperCase() || 'A' }}
                </div>
                <div class="agent-info">
                  <div class="agent-name">{{ agent.name }}</div>
                  <div class="agent-meta">
                    <span class="text-muted text-sm">{{ agent.type }}</span>
                    <span :class="['badge', getStatusClass(agent.status), 'ml-2']">{{ agent.status }}</span>
                  </div>
                </div>
                <div class="agent-load">
                  <div class="load-bar">
                    <div class="load-fill" :style="{ width: agent.load + '%' }"></div>
                  </div>
                  <span class="text-sm text-muted">{{ agent.load }}%</span>
                </div>
              </div>
            </div>
            <div v-else class="empty-state-sm">
              <p class="text-muted">暂无智能体</p>
            </div>
          </div>
        </div>

        <!-- Recent Tasks -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">
              <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
              </svg>
              最近任务
            </h3>
            <router-link to="/tasks" class="btn btn-outline btn-sm">查看全部</router-link>
          </div>
          <div class="card-body">
            <div v-if="tasks?.length" class="task-list">
              <div v-for="task in tasks.slice(0, 5)" :key="task.id" class="task-item">
                <div class="task-status" :class="task.status">
                  <svg v-if="task.status === 'completed'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"/>
                  </svg>
                  <svg v-else-if="task.status === 'failed'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <line x1="15" y1="9" x2="9" y2="15"/>
                    <line x1="9" y1="9" x2="15" y2="15"/>
                  </svg>
                  <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="12" cy="12" r="10"/>
                    <polyline points="12 6 12 12 16 14"/>
                  </svg>
                </div>
                <div class="task-info">
                  <div class="task-skill">{{ task.skillId }}</div>
                  <div class="task-meta">
                    <span class="text-muted text-sm">{{ formatTime(task.createdAt) }}</span>
                    <span v-if="task.duration" class="text-sm ml-2">{{ task.duration }}ms</span>
                  </div>
                </div>
                <span :class="['badge', getStatusClass(task.status)]">{{ task.status }}</span>
              </div>
            </div>
            <div v-else class="empty-state-sm">
              <p class="text-muted">暂无任务</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Performance Section -->
      <div class="card mt-4">
        <div class="card-header">
          <h3 class="card-title">
            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="20" x2="18" y2="10"/>
              <line x1="12" y1="20" x2="12" y2="4"/>
              <line x1="6" y1="20" x2="6" y2="14"/>
            </svg>
            性能指标
          </h3>
        </div>
        <div class="card-body">
          <div class="perf-grid">
            <div class="perf-item">
              <div class="perf-label">平均任务耗时</div>
              <div class="perf-value">{{ Math.round(metrics?.performance?.avgTaskDuration || 0) }}<span class="perf-unit">ms</span></div>
            </div>
            <div class="perf-item">
              <div class="perf-label">平均响应时间</div>
              <div class="perf-value">{{ Math.round(metrics?.performance?.avgResponseTime || 0) }}<span class="perf-unit">ms</span></div>
            </div>
            <div class="perf-item">
              <div class="perf-label">吞吐量</div>
              <div class="perf-value">{{ Math.round(metrics?.performance?.throughput || 0) }}<span class="perf-unit">/s</span></div>
            </div>
            <div class="perf-item">
              <div class="perf-label">任务成功率</div>
              <div class="perf-value">
                {{ metrics?.tasks?.total ? Math.round((metrics.tasks.completed / metrics.tasks.total) * 100) : 0 }}<span class="perf-unit">%</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.text-2xl { font-size: 1.5rem; }
.font-bold { font-weight: 700; }
.text-lg { font-size: 1.125rem; }
.text-sm { font-size: 0.875rem; }
.mt-1 { margin-top: 0.25rem; }
.mt-4 { margin-top: 1rem; }
.mt-6 { margin-top: 1.5rem; }
.ml-2 { margin-left: 0.5rem; }

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
}
.status-dot.connected { background: var(--success); }
.status-dot.disconnected { background: var(--danger); }

.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid var(--gray-200);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Stat Cards with Gradients */
.stat-card {
  display: flex;
  align-items: flex-start;
  gap: 1rem;
  padding: 1.25rem;
  border-radius: 12px;
  color: white;
  position: relative;
  overflow: hidden;
}

.stat-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  opacity: 0.1;
  background: url("data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%23ffffff' fill-opacity='0.4'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E");
}

.gradient-blue { background: linear-gradient(135deg, #3B82F6 0%, #1D4ED8 100%); }
.gradient-green { background: linear-gradient(135deg, #22C55E 0%, #16A34A 100%); }
.gradient-purple { background: linear-gradient(135deg, #8B5CF6 0%, #6D28D9 100%); }
.gradient-orange { background: linear-gradient(135deg, #F59E0B 0%, #D97706 100%); }

.stat-icon {
  width: 48px;
  height: 48px;
  background: rgba(255,255,255,0.2);
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.stat-content { flex: 1; }
.stat-label { font-size: 0.875rem; opacity: 0.9; }
.stat-value { font-size: 1.75rem; font-weight: 700; margin-top: 0.25rem; }
.stat-detail { font-size: 0.75rem; margin-top: 0.25rem; opacity: 0.9; }

/* Dashboard Grid */
.dashboard-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1.5rem;
}

@media (max-width: 1024px) {
  .dashboard-grid { grid-template-columns: 1fr; }
}

/* Card */
.card { background: white; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--gray-100);
}
.card-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
}
.card-body { padding: 1rem 1.25rem; }

/* Agent List */
.agent-list { display: flex; flex-direction: column; gap: 0.75rem; }
.agent-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--gray-50);
  border-radius: 8px;
  transition: background 0.2s;
}
.agent-item:hover { background: var(--gray-100); }

.agent-avatar {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: white;
  background: var(--gray-300);
}
.agent-avatar.online { background: var(--success); }
.agent-avatar.busy { background: var(--warning); }
.agent-avatar.offline { background: var(--gray-400); }

.agent-info { flex: 1; min-width: 0; }
.agent-name { font-weight: 500; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.agent-meta { display: flex; align-items: center; margin-top: 0.125rem; }

.agent-load { display: flex; align-items: center; gap: 0.5rem; min-width: 80px; }
.load-bar { flex: 1; height: 4px; background: var(--gray-200); border-radius: 2px; overflow: hidden; }
.load-fill { height: 100%; background: var(--primary); border-radius: 2px; }

/* Task List */
.task-list { display: flex; flex-direction: column; gap: 0.5rem; }
.task-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--gray-50);
  border-radius: 8px;
}

.task-status {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--gray-200);
  color: var(--gray-500);
}
.task-status.completed { background: #DCFCE7; color: var(--success); }
.task-status.failed { background: #FEE2E2; color: var(--danger); }
.task-status.running { background: #FEF3C7; color: var(--warning); }

.task-info { flex: 1; min-width: 0; }
.task-skill { font-weight: 500; font-size: 0.875rem; }
.task-meta { margin-top: 0.125rem; }

/* Performance Grid */
.perf-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

@media (max-width: 768px) {
  .perf-grid { grid-template-columns: repeat(2, 1fr); }
}

.perf-item {
  text-align: center;
  padding: 1rem;
  background: var(--gray-50);
  border-radius: 8px;
}

.perf-label {
  font-size: 0.75rem;
  color: var(--gray-500);
  margin-bottom: 0.5rem;
}

.perf-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--gray-900);
}

.perf-unit {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--gray-500);
  margin-left: 0.125rem;
}

/* Empty State */
.empty-state-sm {
  text-align: center;
  padding: 2rem;
  color: var(--gray-400);
}
</style>