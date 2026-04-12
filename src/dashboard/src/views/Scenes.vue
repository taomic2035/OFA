<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { apiClient, wsClient } from '../api/client'

// Scene types from Center
const sceneTypes = [
  { id: 'running', name: '跑步', icon: '🏃', color: '#22C55E' },
  { id: 'meeting', name: '会议', icon: '📅', color: '#3B82F6' },
  { id: 'health_alert', name: '健康异常', icon: '⚠️', color: '#EF4444' },
  { id: 'driving', name: '驾驶', icon: '🚗', color: '#F59E0B' },
  { id: 'exercise', name: '运动', icon: '💪', color: '#8B5CF6' },
  { id: 'sleeping', name: '睡眠', icon: '😴', color: '#6366F1' },
  { id: 'work', name: '工作', icon: '💼', color: '#0EA5E9' },
  { id: 'home', name: '家庭', icon: '🏠', color: '#14B8A6' }
]

// State
const activeScenes = ref<any[]>([])
const sceneHistory = ref<any[]>([])
const sceneStats = ref<any>({})
const loading = ref(true)
const selectedScene = ref<string | null>(null)
const isConnected = ref(false)

// Computed
const currentActiveScene = computed(() => {
  if (activeScenes.value.length === 0) return null
  // Return highest priority scene
  return activeScenes.value.sort((a, b) => b.confidence - a.confidence)[0]
})

const sceneCounts = computed(() => {
  const counts: Record<string, number> = {}
  for (const type of sceneTypes) {
    counts[type.id] = sceneHistory.value.filter(s => s.type === type.id).length
  }
  return counts
})

// Methods
async function loadScenes() {
  try {
    // Fetch active scenes
    const activeResponse = await fetch('/api/v1/scene/active')
    if (activeResponse.ok) {
      activeScenes.value = await activeResponse.json()
    }

    // Fetch scene history
    const historyResponse = await fetch('/api/v1/scene/history?limit=50')
    if (historyResponse.ok) {
      sceneHistory.value = await historyResponse.json()
    }

    // Fetch scene stats
    const statsResponse = await fetch('/api/v1/scene/stats')
    if (statsResponse.ok) {
      sceneStats.value = await statsResponse.json()
    }
  } catch (e) {
    console.error('Failed to load scenes:', e)
  } finally {
    loading.value = false
  }
}

function getSceneTypeInfo(typeId: string) {
  return sceneTypes.find(t => t.id === typeId) || { name: typeId, icon: '❓', color: '#6B7280' }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleTimeString()
}

function formatDuration(duration: number): string {
  if (!duration) return '-'
  const mins = Math.floor(duration / 60)
  const secs = duration % 60
  if (mins > 0) return `${mins}分${secs}秒`
  return `${secs}秒`
}

function selectScene(sceneId: string) {
  selectedScene.value = selectedScene.value === sceneId ? null : sceneId
}

// WebSocket handlers
function handleSceneStart(data: any) {
  activeScenes.value.push(data)
}

function handleSceneEnd(data: any) {
  activeScenes.value = activeScenes.value.filter(s => s.type !== data.type)
  sceneHistory.value.unshift(data)
  if (sceneHistory.value.length > 100) {
    sceneHistory.value = sceneHistory.value.slice(0, 100)
  }
}

function handleSceneAction(data: any) {
  // Update scene with action info
  const scene = activeScenes.value.find(s => s.type === data.sceneType)
  if (scene) {
    if (!scene.actionsExecuted) scene.actionsExecuted = []
    scene.actionsExecuted.push(data.action)
  }
}

onMounted(async () => {
  await loadScenes()

  // WebSocket connection
  wsClient.connect()
  wsClient.on('connected', () => { isConnected.value = true })
  wsClient.on('disconnected', () => { isConnected.value = false })
  wsClient.on('SceneStart', handleSceneStart)
  wsClient.on('SceneEnd', handleSceneEnd)
  wsClient.on('SceneAction', handleSceneAction)

  // Auto refresh every 10 seconds
  const interval = setInterval(loadScenes, 10000)

  onUnmounted(() => {
    clearInterval(interval)
    wsClient.disconnect()
  })
})
</script>

<template>
  <div class="scene-view">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>场景检测</h1>
        <p class="text-muted">实时场景状态和历史记录</p>
      </div>
      <div class="header-actions">
        <div class="connection-indicator">
          <span :class="['status-dot', isConnected ? 'connected' : 'disconnected']"></span>
          {{ isConnected ? '实时更新' : '离线' }}
        </div>
        <button class="btn btn-outline" @click="loadScenes" :disabled="loading">
          刷新
        </button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载场景数据...</p>
    </div>

    <!-- Content -->
    <template v-else>
      <!-- Current Scene -->
      <div class="current-scene-section">
        <h2>当前活跃场景</h2>
        <div v-if="currentActiveScene" class="active-scene-card">
          <div class="scene-icon-large" :style="{ background: getSceneTypeInfo(currentActiveScene.type).color }">
            {{ getSceneTypeInfo(currentActiveScene.type).icon }}
          </div>
          <div class="scene-details">
            <div class="scene-name">{{ getSceneTypeInfo(currentActiveScene.type).name }}</div>
            <div class="scene-meta">
              <span class="confidence">置信度: {{ (currentActiveScene.confidence * 100).toFixed(0) }}%</span>
              <span class="duration">持续: {{ formatTime(currentActiveScene.startTime) }}</span>
            </div>
            <div class="scene-agent">设备: {{ currentActiveScene.agentId }}</div>
          </div>
          <div class="scene-actions">
            <div class="actions-header">待执行动作:</div>
            <div class="actions-list">
              <div v-for="action in currentActiveScene.actions" :key="action.id" class="action-item">
                <span :class="['action-status', action.executed ? 'done' : 'pending']">
                  {{ action.executed ? '✓' : '○' }}
                </span>
                <span class="action-type">{{ action.type }}</span>
                <span class="action-target">{{ action.targetAgent || '全部' }}</span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="no-active-scene">
          <div class="empty-icon">🎬</div>
          <p>当前没有活跃场景</p>
        </div>
      </div>

      <!-- Scene Type Summary -->
      <div class="scene-summary">
        <h2>场景统计</h2>
        <div class="scene-types-grid">
          <div v-for="type in sceneTypes" :key="type.id"
               :class="['scene-type-card', { active: activeScenes.value.some(s => s.type === type.id) }]"
               :style="{ borderColor: type.color }">
            <div class="type-icon" :style="{ background: type.color }">
              {{ type.icon }}
            </div>
            <div class="type-info">
              <div class="type-name">{{ type.name }}</div>
              <div class="type-count">{{ sceneCounts[type.id] || 0 }} 次</div>
            </div>
            <div v-if="activeScenes.value.some(s => s.type === type.id)" class="active-indicator">
              活跃
            </div>
          </div>
        </div>
      </div>

      <!-- Scene History -->
      <div class="scene-history">
        <h2>场景历史 (最近50条)</h2>
        <div v-if="sceneHistory.length" class="history-list">
          <div v-for="scene in sceneHistory" :key="scene.startTime"
               class="history-item"
               @click="selectScene(scene.startTime)">
            <div class="history-icon" :style="{ background: getSceneTypeInfo(scene.type).color }">
              {{ getSceneTypeInfo(scene.type).icon }}
            </div>
            <div class="history-info">
              <div class="history-type">{{ getSceneTypeInfo(scene.type).name }}</div>
              <div class="history-meta">
                <span>{{ formatTime(scene.startTime) }}</span>
                <span v-if="scene.endTime"> - {{ formatTime(scene.endTime) }}</span>
                <span v-if="scene.duration">({{ formatDuration(scene.duration) }})</span>
              </div>
            </div>
            <div class="history-confidence">
              {{ (scene.confidence * 100).toFixed(0) }}%
            </div>

            <!-- Expanded Details -->
            <div v-if="selectedScene === scene.startTime" class="history-details">
              <div class="detail-section">
                <h4>场景上下文</h4>
                <div class="context-grid">
                  <div v-for="(value, key) in scene.context" :key="key" class="context-item">
                    <span class="context-key">{{ key }}</span>
                    <span class="context-value">{{ value }}</span>
                  </div>
                </div>
              </div>
              <div v-if="scene.actions" class="detail-section">
                <h4>执行动作</h4>
                <div class="actions-executed">
                  <div v-for="action in scene.actions" :key="action.id" class="executed-action">
                    <span :class="['exec-status', action.executed ? 'success' : 'pending']">
                      {{ action.executed ? '✓' : '○' }}
                    </span>
                    <span>{{ action.type }} → {{ action.targetAgent || '全部' }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>暂无场景历史记录</p>
        </div>
      </div>

      <!-- Scene Engine Stats -->
      <div class="engine-stats">
        <h2>引擎统计</h2>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ sceneStats.activeScenes || 0 }}</div>
            <div class="stat-label">活跃场景数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ sceneStats.historyCount || 0 }}</div>
            <div class="stat-label">历史记录数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ sceneStats.ruleCount || 0 }}</div>
            <div class="stat-label">触发规则数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ sceneStats.detectorCount || 8 }}</div>
            <div class="stat-label">检测器数</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ (sceneStats.minConfidence || 0.7) * 100 }}%</div>
            <div class="stat-label">最小置信度</div>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.scene-view { padding: 0; }

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
}

.page-header h1 { margin: 0; font-size: 1.5rem; }
.page-header p { margin: 0.25rem 0 0 0; }

.header-actions { display: flex; gap: 1rem; align-items: center; }

.connection-indicator {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.875rem;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
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

@keyframes spin { to { transform: rotate(360deg); } }

h2 {
  font-size: 1.125rem;
  margin: 0 0 1rem 0;
  color: var(--gray-800);
}

/* Current Scene */
.current-scene-section { margin-bottom: 2rem; }

.active-scene-card {
  display: flex;
  gap: 1.5rem;
  padding: 1.5rem;
  background: linear-gradient(135deg, #F0F9FF 0%, #E0F2FE 100%);
  border-radius: 16px;
  border: 2px solid var(--primary);
}

.scene-icon-large {
  width: 64px;
  height: 64px;
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  color: white;
}

.scene-details { flex: 1; }

.scene-name { font-size: 1.5rem; font-weight: 700; margin-bottom: 0.5rem; }

.scene-meta { display: flex; gap: 1rem; margin-bottom: 0.25rem; }
.confidence { color: var(--success); font-weight: 500; }
.duration { color: var(--gray-500); }

.scene-agent { font-size: 0.875rem; color: var(--gray-600); }

.scene-actions { min-width: 200px; }

.actions-header { font-size: 0.75rem; color: var(--gray-500); margin-bottom: 0.5rem; }

.actions-list { display: flex; flex-direction: column; gap: 0.25rem; }

.action-item { display: flex; gap: 0.5rem; font-size: 0.875rem; }
.action-status.done { color: var(--success); }
.action-status.pending { color: var(--gray-400); }
.action-type { font-weight: 500; }
.action-target { color: var(--gray-500); }

.no-active-scene {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  background: var(--gray-50);
  border-radius: 16px;
  text-align: center;
}

.empty-icon { font-size: 3rem; margin-bottom: 1rem; opacity: 0.5; }

/* Scene Summary */
.scene-summary { margin-bottom: 2rem; }

.scene-types-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

@media (max-width: 768px) {
  .scene-types-grid { grid-template-columns: repeat(2, 1fr); }
}

.scene-type-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: white;
  border-radius: 12px;
  border: 2px solid transparent;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  transition: all 0.2s;
}

.scene-type-card.active {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.1) 0%, rgba(34, 197, 94, 0.05) 100%);
}

.type-icon {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  color: white;
}

.type-info { flex: 1; }
.type-name { font-weight: 600; }
.type-count { font-size: 0.875rem; color: var(--gray-500); }

.active-indicator {
  padding: 0.25rem 0.75rem;
  background: var(--success);
  color: white;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
}

/* Scene History */
.scene-history { margin-bottom: 2rem; }

.history-list { display: flex; flex-direction: column; gap: 0.5rem; }

.history-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  cursor: pointer;
  transition: all 0.2s;
}

.history-item:hover { background: var(--gray-50); }

.history-icon {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  color: white;
}

.history-info { flex: 1; }
.history-type { font-weight: 500; }
.history-meta { font-size: 0.875rem; color: var(--gray-500); margin-top: 0.25rem; }

.history-confidence {
  padding: 0.25rem 0.75rem;
  background: var(--gray-100);
  border-radius: 8px;
  font-size: 0.875rem;
  font-weight: 500;
}

.history-details {
  margin-top: 1rem;
  padding-top: 1rem;
  border-top: 1px solid var(--gray-100);
}

.detail-section { margin-bottom: 1rem; }

.detail-section h4 {
  font-size: 0.875rem;
  margin: 0 0 0.5rem 0;
  color: var(--gray-600);
}

.context-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.5rem;
}

@media (max-width: 768px) {
  .context-grid { grid-template-columns: repeat(2, 1fr); }
}

.context-item {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem;
  background: var(--gray-50);
  border-radius: 6px;
}

.context-key { font-size: 0.75rem; color: var(--gray-500); }
.context-value { font-size: 0.75rem; font-weight: 500; }

.actions-executed { display: flex; flex-direction: column; gap: 0.25rem; }

.executed-action { display: flex; gap: 0.5rem; font-size: 0.875rem; }
.exec-status.success { color: var(--success); }
.exec-status.pending { color: var(--gray-400); }

.empty-state { padding: 2rem; text-align: center; color: var(--gray-400); }

/* Engine Stats */
.engine-stats { margin-bottom: 1rem; }

.stats-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1rem;
}

@media (max-width: 768px) {
  .stats-grid { grid-template-columns: repeat(3, 1fr); }
}

.stat-item {
  text-align: center;
  padding: 1rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.stat-value { font-size: 1.5rem; font-weight: 700; color: var(--primary); }
.stat-label { font-size: 0.75rem; color: var(--gray-500); margin-top: 0.25rem; }
</style>