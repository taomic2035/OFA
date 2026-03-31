<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { apiClient } from '../api/client'
import type { Agent } from '../types'

const agents = ref<Agent[]>([])
const loading = ref(true)
const searchQuery = ref('')
const statusFilter = ref('')
const selectedAgent = ref<Agent | null>(null)

async function loadAgents() {
  loading.value = true
  try {
    const response = await apiClient.getAgents({
      status: statusFilter.value || undefined
    })
    agents.value = response.agents || []
  } catch (e) {
    console.error('Failed to load agents:', e)
  } finally {
    loading.value = false
  }
}

async function deleteAgent(id: string) {
  if (!confirm('确定要删除此智能体吗？')) return
  try {
    await apiClient.deleteAgent(id)
    agents.value = agents.value.filter(a => a.id !== id)
  } catch (e) {
    console.error('Failed to delete agent:', e)
    alert('删除失败')
  }
}

const filteredAgents = computed(() => {
  let result = agents.value
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(a =>
      a.id.toLowerCase().includes(query) ||
      a.name.toLowerCase().includes(query) ||
      a.type.toLowerCase().includes(query)
    )
  }
  if (statusFilter.value) {
    result = result.filter(a => a.status === statusFilter.value)
  }
  return result
})

function getStatusClass(status: string): string {
  switch (status) {
    case 'online': return 'status-online'
    case 'offline': return 'status-offline'
    case 'busy': return 'status-busy'
    default: return 'status-idle'
  }
}

function getStatusBadge(status: string): string {
  switch (status) {
    case 'online': return 'badge-success'
    case 'offline': return 'badge-danger'
    case 'busy': return 'badge-warning'
    default: return 'badge-gray'
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

onMounted(loadAgents)
</script>

<template>
  <div>
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>智能体管理</h1>
        <p class="text-muted">管理所有注册的智能体节点</p>
      </div>
      <button class="btn btn-outline" @click="loadAgents" :disabled="loading">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
        刷新
      </button>
    </div>

    <!-- Filters -->
    <div class="filters-bar">
      <div class="search-box">
        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="11" cy="11" r="8"/>
          <line x1="21" y1="21" x2="16.65" y2="16.65"/>
        </svg>
        <input type="text" v-model="searchQuery" placeholder="搜索智能体..." class="search-input">
      </div>
      <div class="filter-tabs">
        <button :class="['filter-tab', { active: statusFilter === '' }]" @click="statusFilter = ''">全部</button>
        <button :class="['filter-tab', { active: statusFilter === 'online' }]" @click="statusFilter = 'online'">在线</button>
        <button :class="['filter-tab', { active: statusFilter === 'busy' }]" @click="statusFilter = 'busy'">忙碌</button>
        <button :class="['filter-tab', { active: statusFilter === 'offline' }]" @click="statusFilter = 'offline'">离线</button>
      </div>
    </div>

    <!-- Stats -->
    <div class="quick-stats">
      <div class="quick-stat">
        <span class="quick-stat-value">{{ agents.length }}</span>
        <span class="quick-stat-label">总数</span>
      </div>
      <div class="quick-stat">
        <span class="quick-stat-value text-success">{{ agents.filter(a => a.status === 'online').length }}</span>
        <span class="quick-stat-label">在线</span>
      </div>
      <div class="quick-stat">
        <span class="quick-stat-value text-warning">{{ agents.filter(a => a.status === 'busy').length }}</span>
        <span class="quick-stat-label">忙碌</span>
      </div>
      <div class="quick-stat">
        <span class="quick-stat-value text-danger">{{ agents.filter(a => a.status === 'offline').length }}</span>
        <span class="quick-stat-label">离线</span>
      </div>
    </div>

    <!-- Agent Grid -->
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
    </div>

    <div v-else-if="filteredAgents.length === 0" class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
        <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/>
        <circle cx="9" cy="7" r="4"/>
      </svg>
      <h3>暂无智能体</h3>
      <p class="text-muted">没有找到匹配的智能体</p>
    </div>

    <div v-else class="agent-grid">
      <div v-for="agent in filteredAgents" :key="agent.id" class="agent-card" @click="selectedAgent = agent">
        <div class="agent-card-header">
          <div :class="['agent-avatar', getStatusClass(agent.status)]">
            {{ agent.name?.charAt(0)?.toUpperCase() || 'A' }}
          </div>
          <div class="agent-info">
            <div class="agent-name">{{ agent.name }}</div>
            <div class="agent-type">{{ agent.type }}</div>
          </div>
          <span :class="['status-badge', getStatusBadge(agent.status)]">{{ agent.status }}</span>
        </div>

        <div class="agent-card-body">
          <div class="info-row">
            <span class="info-label">ID</span>
            <span class="info-value"><code>{{ agent.id?.slice(0, 16) }}...</code></span>
          </div>
          <div class="info-row">
            <span class="info-label">负载</span>
            <div class="load-indicator">
              <div class="load-bar">
                <div class="load-fill" :style="{ width: agent.load + '%' }"></div>
              </div>
              <span>{{ agent.load }}%</span>
            </div>
          </div>
          <div class="info-row">
            <span class="info-label">能力</span>
            <span class="info-value">{{ agent.capabilities?.length || 0 }} 项技能</span>
          </div>
        </div>

        <div class="agent-card-footer">
          <span class="text-muted text-sm">最后在线: {{ formatTime(agent.lastSeen) }}</span>
          <button class="btn btn-danger btn-sm" @click.stop="deleteAgent(agent.id)">删除</button>
        </div>
      </div>
    </div>

    <!-- Agent Detail Modal -->
    <div v-if="selectedAgent" class="modal-overlay" @click="selectedAgent = null">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>{{ selectedAgent.name }}</h3>
          <button class="modal-close" @click="selectedAgent = null">×</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h4>基本信息</h4>
            <div class="detail-grid">
              <div class="detail-item">
                <span class="detail-label">ID</span>
                <code class="detail-value">{{ selectedAgent.id }}</code>
              </div>
              <div class="detail-item">
                <span class="detail-label">类型</span>
                <span class="detail-value">{{ selectedAgent.type }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">状态</span>
                <span :class="['badge', getStatusBadge(selectedAgent.status)]">{{ selectedAgent.status }}</span>
              </div>
              <div class="detail-item">
                <span class="detail-label">负载</span>
                <span class="detail-value">{{ selectedAgent.load }}%</span>
              </div>
            </div>
          </div>

          <div class="detail-section">
            <h4>能力列表</h4>
            <div v-if="selectedAgent.capabilities?.length" class="capabilities-list">
              <div v-for="cap in selectedAgent.capabilities" :key="cap.id" class="capability-item">
                <div class="capability-header">
                  <span class="capability-name">{{ cap.name }}</span>
                  <span class="capability-id">{{ cap.id }}</span>
                </div>
                <div class="capability-operations">
                  <span v-for="op in cap.operations" :key="op" class="operation-tag">{{ op }}</span>
                </div>
              </div>
            </div>
            <div v-else class="text-muted">暂无能力</div>
          </div>
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

.filters-bar {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.search-box {
  position: relative;
  flex: 1;
  min-width: 250px;
}

.search-box svg {
  position: absolute;
  left: 12px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--gray-400);
}

.search-input {
  width: 100%;
  padding: 0.625rem 0.875rem 0.625rem 2.5rem;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-size: 0.875rem;
}

.search-input:focus {
  outline: none;
  border-color: var(--primary);
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

.filter-tabs {
  display: flex;
  background: var(--gray-100);
  border-radius: 8px;
  padding: 4px;
}

.filter-tab {
  padding: 0.5rem 1rem;
  border: none;
  background: transparent;
  border-radius: 6px;
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s;
}

.filter-tab:hover { background: var(--gray-200); }
.filter-tab.active { background: white; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }

.quick-stats {
  display: flex;
  gap: 2rem;
  margin-bottom: 1.5rem;
  padding: 1rem 1.5rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.quick-stat { display: flex; flex-direction: column; }
.quick-stat-value { font-size: 1.5rem; font-weight: 700; }
.quick-stat-label { font-size: 0.75rem; color: var(--gray-500); margin-top: 0.25rem; }

.agent-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1rem;
}

.agent-card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
}

.agent-card:hover {
  box-shadow: 0 4px 12px rgba(0,0,0,0.15);
  transform: translateY(-2px);
}

.agent-card-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  border-bottom: 1px solid var(--gray-100);
}

.agent-avatar {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 1.125rem;
  color: white;
}

.agent-avatar.status-online { background: var(--success); }
.agent-avatar.status-busy { background: var(--warning); }
.agent-avatar.status-offline { background: var(--gray-400); }
.agent-avatar.status-idle { background: var(--gray-300); }

.agent-info { flex: 1; min-width: 0; }
.agent-name { font-weight: 600; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.agent-type { font-size: 0.75rem; color: var(--gray-500); }

.status-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.agent-card-body { padding: 1rem; }

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
}

.info-row:not(:last-child) { border-bottom: 1px solid var(--gray-100); }

.info-label { font-size: 0.75rem; color: var(--gray-500); }
.info-value { font-size: 0.875rem; }

.load-indicator { display: flex; align-items: center; gap: 0.5rem; }
.load-bar { width: 60px; height: 4px; background: var(--gray-200); border-radius: 2px; overflow: hidden; }
.load-fill { height: 100%; background: var(--primary); border-radius: 2px; }

.agent-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: var(--gray-50);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 16px;
  width: 600px;
  max-width: 90%;
  max-height: 85vh;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--gray-100);
}

.modal-header h3 { margin: 0; }
.modal-close { width: 32px; height: 32px; border: none; background: var(--gray-100); border-radius: 8px; font-size: 1.25rem; cursor: pointer; }
.modal-body { padding: 1.5rem; overflow-y: auto; max-height: calc(85vh - 70px); }

.detail-section { margin-bottom: 1.5rem; }
.detail-section:last-child { margin-bottom: 0; }
.detail-section h4 { margin: 0 0 0.75rem 0; font-size: 0.875rem; color: var(--gray-500); text-transform: uppercase; letter-spacing: 0.05em; }

.detail-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 1rem; }
.detail-item { display: flex; flex-direction: column; gap: 0.25rem; }
.detail-label { font-size: 0.75rem; color: var(--gray-500); }
.detail-value { font-size: 0.875rem; }

.capabilities-list { display: flex; flex-direction: column; gap: 0.75rem; }
.capability-item { padding: 0.75rem; background: var(--gray-50); border-radius: 8px; }
.capability-header { display: flex; justify-content: space-between; margin-bottom: 0.5rem; }
.capability-name { font-weight: 500; }
.capability-id { font-size: 0.75rem; color: var(--gray-400); }
.capability-operations { display: flex; flex-wrap: wrap; gap: 0.25rem; }
.operation-tag { padding: 0.125rem 0.5rem; background: white; border-radius: 4px; font-size: 0.75rem; }

.loading-state { display: flex; justify-content: center; padding: 4rem; }
.loading-spinner { width: 40px; height: 40px; border: 3px solid var(--gray-200); border-top-color: var(--primary); border-radius: 50%; animation: spin 1s linear infinite; }

.empty-state { text-align: center; padding: 4rem; }
.empty-state h3 { margin: 1rem 0 0.5rem 0; }
</style>