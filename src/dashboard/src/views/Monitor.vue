<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { apiClient, wsClient } from '../api/client'
import type { SystemMetrics } from '../types'

const metrics = ref<SystemMetrics | null>(null)
const systemInfo = ref<any>(null)
const loading = ref(true)
const autoRefresh = ref(true)
const isConnected = ref(false)

let refreshInterval: ReturnType<typeof setInterval> | null = null

async function loadMetrics() {
  try {
    const [metricsData, info] = await Promise.all([
      apiClient.getMetrics(),
      apiClient.getSystemInfo()
    ])
    metrics.value = metricsData
    systemInfo.value = info
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

function formatUptime(seconds: number): string {
  if (!seconds) return '0分钟'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}天 ${hours}小时`
  if (hours > 0) return `${hours}小时 ${mins}分钟`
  return `${mins}分钟`
}

onMounted(() => {
  loadMetrics()
  if (autoRefresh.value) {
    refreshInterval = setInterval(loadMetrics, 5000)
  }
  wsClient.connect()
  wsClient.on('connected', () => { isConnected.value = true })
  wsClient.on('disconnected', () => { isConnected.value = false })
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
  wsClient.disconnect()
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>系统监控</h1>
        <p class="text-muted">实时系统指标和性能数据</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-outline" @click="toggleAutoRefresh">
          <span :class="['indicator', autoRefresh ? 'on' : 'off']"></span>
          {{ autoRefresh ? '自动刷新' : '手动刷新' }}
        </button>
        <button class="btn btn-outline" @click="loadMetrics" :disabled="loading">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
          刷新
        </button>
      </div>
    </div>

    <!-- System Info Banner -->
    <div class="system-banner" v-if="systemInfo">
      <div class="banner-item">
        <span class="banner-label">版本</span>
        <span class="banner-value">v{{ systemInfo.version }}</span>
      </div>
      <div class="banner-item">
        <span class="banner-label">运行时间</span>
        <span class="banner-value">{{ formatUptime(systemInfo.uptime) }}</span>
      </div>
      <div class="banner-item">
        <span class="banner-label">连接状态</span>
        <span :class="['connection-status', isConnected ? 'connected' : 'disconnected']">
          {{ isConnected ? '已连接' : '未连接' }}
        </span>
      </div>
    </div>

    <!-- Main Metrics -->
    <div class="metrics-grid">
      <!-- Agents -->
      <div class="metric-card">
        <div class="metric-header">
          <div class="metric-icon blue">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
              <circle cx="9" cy="7" r="4"/>
            </svg>
          </div>
          <h3>智能体</h3>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ metrics?.agents?.total || 0 }}</span>
            <span class="metric-label">总数</span>
          </div>
          <div class="metric-details">
            <div class="detail-item">
              <span class="detail-dot success"></span>
              <span class="detail-text">在线: {{ metrics?.agents?.online || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-dot warning"></span>
              <span class="detail-text">忙碌: {{ metrics?.agents?.busy || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-dot danger"></span>
              <span class="detail-text">离线: {{ metrics?.agents?.offline || 0 }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Tasks -->
      <div class="metric-card">
        <div class="metric-header">
          <div class="metric-icon green">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
            </svg>
          </div>
          <h3>任务</h3>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ metrics?.tasks?.total || 0 }}</span>
            <span class="metric-label">总数</span>
          </div>
          <div class="metric-details">
            <div class="detail-item">
              <span class="detail-dot success"></span>
              <span class="detail-text">完成: {{ metrics?.tasks?.completed || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-dot warning"></span>
              <span class="detail-text">运行: {{ metrics?.tasks?.running || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-dot info"></span>
              <span class="detail-text">等待: {{ metrics?.tasks?.pending || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-dot danger"></span>
              <span class="detail-text">失败: {{ metrics?.tasks?.failed || 0 }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Messages -->
      <div class="metric-card">
        <div class="metric-header">
          <div class="metric-icon purple">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <h3>消息</h3>
        </div>
        <div class="metric-body">
          <div class="metric-main">
            <span class="metric-value">{{ metrics?.messages?.total || 0 }}</span>
            <span class="metric-label">总数</span>
          </div>
          <div class="metric-details">
            <div class="detail-item">
              <span class="detail-dot success"></span>
              <span class="detail-text">已送达: {{ metrics?.messages?.delivered || 0 }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-dot info"></span>
              <span class="detail-text">等待中: {{ metrics?.messages?.pending || 0 }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Performance -->
      <div class="metric-card">
        <div class="metric-header">
          <div class="metric-icon orange">
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="20" x2="18" y2="10"/>
              <line x1="12" y1="20" x2="12" y2="4"/>
              <line x1="6" y1="20" x2="6" y2="14"/>
            </svg>
          </div>
          <h3>性能</h3>
        </div>
        <div class="metric-body">
          <div class="perf-grid">
            <div class="perf-item">
              <span class="perf-value">{{ Math.round(metrics?.performance?.avgTaskDuration || 0) }}</span>
              <span class="perf-label">平均耗时(ms)</span>
            </div>
            <div class="perf-item">
              <span class="perf-value">{{ Math.round(metrics?.performance?.avgResponseTime || 0) }}</span>
              <span class="perf-label">响应时间(ms)</span>
            </div>
            <div class="perf-item">
              <span class="perf-value">{{ Math.round(metrics?.performance?.throughput || 0) }}</span>
              <span class="perf-label">吞吐量(/s)</span>
            </div>
            <div class="perf-item">
              <span class="perf-value">{{ metrics?.tasks?.total ? Math.round((metrics.tasks.completed / metrics.tasks.total) * 100) : 0 }}%</span>
              <span class="perf-label">成功率</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Task Progress -->
    <div class="progress-section">
      <h3>任务进度</h3>
      <div class="progress-bar-container">
        <div class="progress-bar">
          <div class="progress-fill completed" :style="{ width: `${metrics?.tasks?.total ? (metrics.tasks.completed / metrics.tasks.total) * 100 : 0}%` }"></div>
          <div class="progress-fill running" :style="{ width: `${metrics?.tasks?.total ? (metrics.tasks.running / metrics.tasks.total) * 100 : 0}%` }"></div>
          <div class="progress-fill failed" :style="{ width: `${metrics?.tasks?.total ? (metrics.tasks.failed / metrics.tasks.total) * 100 : 0}%` }"></div>
        </div>
        <div class="progress-legend">
          <span class="legend-item"><span class="legend-color completed"></span> 完成</span>
          <span class="legend-item"><span class="legend-color running"></span> 运行</span>
          <span class="legend-item"><span class="legend-color failed"></span> 失败</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 { margin: 0; font-size: 1.5rem; }
.page-header p { margin: 0.25rem 0 0 0; }

.header-actions { display: flex; gap: 0.5rem; }

.indicator {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
  margin-right: 0.5rem;
}
.indicator.on { background: var(--success); }
.indicator.off { background: var(--gray-400); }

.system-banner {
  display: flex;
  gap: 2rem;
  padding: 1rem 1.5rem;
  background: linear-gradient(135deg, #3B82F6 0%, #1D4ED8 100%);
  border-radius: 12px;
  color: white;
  margin-bottom: 1.5rem;
}

.banner-item { display: flex; flex-direction: column; gap: 0.25rem; }
.banner-label { font-size: 0.75rem; opacity: 0.8; }
.banner-value { font-weight: 600; }

.connection-status {
  padding: 0.25rem 0.75rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
}
.connection-status.connected { background: rgba(255,255,255,0.2); }
.connection-status.disconnected { background: rgba(239,68,68,0.5); }

.metrics-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

@media (max-width: 768px) {
  .metrics-grid { grid-template-columns: 1fr; }
}

.metric-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.metric-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 1.25rem;
  background: var(--gray-50);
  border-bottom: 1px solid var(--gray-100);
}

.metric-header h3 { margin: 0; font-size: 1rem; }

.metric-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
}
.metric-icon.blue { background: #DBEAFE; color: var(--primary); }
.metric-icon.green { background: #DCFCE7; color: var(--success); }
.metric-icon.purple { background: #EDE9FE; color: #8B5CF6; }
.metric-icon.orange { background: #FEF3C7; color: var(--warning); }

.metric-body { padding: 1.25rem; }

.metric-main {
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.metric-value { font-size: 2rem; font-weight: 700; }
.metric-label { color: var(--gray-500); }

.metric-details { display: flex; flex-direction: column; gap: 0.5rem; }

.detail-item { display: flex; align-items: center; gap: 0.5rem; }
.detail-dot { width: 8px; height: 8px; border-radius: 50%; }
.detail-dot.success { background: var(--success); }
.detail-dot.warning { background: var(--warning); }
.detail-dot.danger { background: var(--danger); }
.detail-dot.info { background: var(--primary); }
.detail-text { font-size: 0.875rem; }

.perf-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.perf-item {
  text-align: center;
  padding: 0.75rem;
  background: var(--gray-50);
  border-radius: 8px;
}

.perf-value { display: block; font-size: 1.25rem; font-weight: 700; }
.perf-label { display: block; font-size: 0.75rem; color: var(--gray-500); margin-top: 0.25rem; }

.progress-section {
  background: white;
  border-radius: 12px;
  padding: 1.25rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.progress-section h3 { margin: 0 0 1rem 0; font-size: 1rem; }

.progress-bar-container { display: flex; flex-direction: column; gap: 0.75rem; }

.progress-bar {
  height: 24px;
  background: var(--gray-100);
  border-radius: 12px;
  overflow: hidden;
  display: flex;
}

.progress-fill { height: 100%; }
.progress-fill.completed { background: var(--success); }
.progress-fill.running { background: var(--warning); }
.progress-fill.failed { background: var(--danger); }

.progress-legend { display: flex; gap: 1.5rem; }

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
  color: var(--gray-600);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
}
.legend-color.completed { background: var(--success); }
.legend-color.running { background: var(--warning); }
.legend-color.failed { background: var(--danger); }
</style>