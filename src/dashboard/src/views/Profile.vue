<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { apiClient } from '../api/client'
import type { PersonalIdentity, Personality, ValueSystem, Interest, Memory, Preference } from '../types'

const loading = ref(true)
const activeTab = ref<'identity' | 'personality' | 'values' | 'interests' | 'memory' | 'preferences'>('identity')
const userId = ref('user-001')  // Default user ID

// Data
const identity = ref<PersonalIdentity | null>(null)
const memories = ref<Memory[]>([])
const preferences = ref<Preference[]>([])

// Edit states
const editingField = ref<string | null>(null)
const editValue = ref<any>(null)

// Load data
async function loadIdentity() {
  loading.value = true
  try {
    const response = await apiClient.getIdentity(userId.value)
    if (response.success && response.identity) {
      identity.value = response.identity
    }
  } catch (e) {
    console.error('Failed to load identity:', e)
  } finally {
    loading.value = false
  }
}

async function loadMemories() {
  try {
    const response = await apiClient.getMemories(userId.value)
    memories.value = response.memories || []
  } catch (e) {
    console.error('Failed to load memories:', e)
  }
}

async function loadPreferences() {
  try {
    const response = await apiClient.getPreferences(userId.value)
    preferences.value = response.preferences || []
  } catch (e) {
    console.error('Failed to load preferences:', e)
  }
}

// Computed
const mbtiDisplay = computed(() => {
  if (!identity.value?.personality?.mbtiType) return '未设置'
  return identity.value.personality.mbtiType
})

const bigFiveScores = computed(() => {
  const p = identity.value?.personality
  if (!p) return []
  return [
    { name: '开放性', value: (p.openness || 0) * 100, color: '#8B5CF6' },
    { name: '尽责性', value: (p.conscientiousness || 0) * 100, color: '#3B82F6' },
    { name: '外向性', value: (p.extraversion || 0) * 100, color: '#22C55E' },
    { name: '宜人性', value: (p.agreeableness || 0) * 100, color: '#F59E0B' },
    { name: '神经质', value: (p.neuroticism || 0) * 100, color: '#EF4444' }
  ]
})

const valueScores = computed(() => {
  const v = identity.value?.valueSystem
  if (!v) return []
  return [
    { name: '隐私', value: (v.privacy || 0) * 100 },
    { name: '效率', value: (v.efficiency || 0) * 100 },
    { name: '健康', value: (v.health || 0) * 100 },
    { name: '家庭', value: (v.family || 0) * 100 },
    { name: '事业', value: (v.career || 0) * 100 },
    { name: '娱乐', value: (v.entertainment || 0) * 100 },
    { name: '学习', value: (v.learning || 0) * 100 },
    { name: '社交', value: (v.social || 0) * 100 },
    { name: '财务', value: (v.finance || 0) * 100 },
    { name: '环境', value: (v.environment || 0) * 100 }
  ]
})

// Functions
function startEdit(field: string, value: any) {
  editingField.value = field
  editValue.value = value
}

async function saveEdit() {
  if (!editingField.value || editValue.value === null) return

  try {
    const updates = { [editingField.value]: editValue.value }
    await apiClient.updateIdentity(userId.value, updates)
    if (identity.value) {
      (identity.value as any)[editingField.value] = editValue.value
    }
  } catch (e) {
    console.error('Failed to save:', e)
  } finally {
    editingField.value = null
    editValue.value = null
  }
}

function cancelEdit() {
  editingField.value = null
  editValue.value = null
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

function getInterestLevel(level: number): string {
  const levels = ['一般', '感兴趣', '喜欢', '热爱', '狂热']
  return levels[Math.min(level - 1, 4)] || '未知'
}

function getInterestColor(level: number): string {
  const colors = ['#9CA3AF', '#60A5FA', '#34D399', '#FBBF24', '#F87171']
  return colors[Math.min(level - 1, 4)] || '#9CA3AF'
}

onMounted(() => {
  loadIdentity()
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="flex items-center justify-between mb-6">
      <div>
        <h1 class="text-2xl font-bold" style="margin: 0;">用户画像</h1>
        <p class="text-muted mt-1" style="margin: 0;">管理个人身份、偏好和记忆</p>
      </div>
      <button class="btn btn-primary" @click="loadIdentity">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
        刷新
      </button>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="loading-container">
      <div class="loading-spinner"></div>
      <p>加载中...</p>
    </div>

    <!-- Content -->
    <template v-else>
      <!-- Tabs -->
      <div class="tabs-container mb-6">
        <button
          v-for="tab in ['identity', 'personality', 'values', 'interests', 'memory', 'preferences']"
          :key="tab"
          :class="['tab-btn', activeTab === tab ? 'active' : '']"
          @click="activeTab = tab as any"
        >
          {{ { identity: '基本信息', personality: '性格特征', values: '价值观', interests: '兴趣爱好', memory: '记忆库', preferences: '偏好设置' }[tab] }}
        </button>
      </div>

      <!-- Identity Tab -->
      <div v-if="activeTab === 'identity'" class="profile-section">
        <div class="profile-header">
          <div class="avatar-large">
            {{ identity?.name?.charAt(0)?.toUpperCase() || 'U' }}
          </div>
          <div class="profile-info">
            <h2 class="profile-name">
              <template v-if="editingField === 'name'">
                <input v-model="editValue" class="edit-input" @blur="saveEdit" @keyup.enter="saveEdit" @keyup.esc="cancelEdit" />
              </template>
              <template v-else>
                {{ identity?.name || '未设置' }}
                <button class="edit-btn" @click="startEdit('name', identity?.name || '')">编辑</button>
              </template>
            </h2>
            <p class="profile-nickname">{{ identity?.nickname || '暂无昵称' }}</p>
          </div>
        </div>

        <div class="info-grid">
          <div class="info-item">
            <label>性别</label>
            <span>{{ identity?.gender || '未设置' }}</span>
          </div>
          <div class="info-item">
            <label>生日</label>
            <span>{{ formatDate(identity?.birthday) }}</span>
          </div>
          <div class="info-item">
            <label>位置</label>
            <span>{{ identity?.location || '未设置' }}</span>
          </div>
          <div class="info-item">
            <label>职业</label>
            <span>{{ identity?.occupation || '未设置' }}</span>
          </div>
          <div class="info-item">
            <label>语言</label>
            <span>{{ identity?.languages?.join(', ') || '未设置' }}</span>
          </div>
          <div class="info-item">
            <label>时区</label>
            <span>{{ identity?.timezone || '未设置' }}</span>
          </div>
        </div>
      </div>

      <!-- Personality Tab -->
      <div v-if="activeTab === 'personality'" class="profile-section">
        <!-- MBTI -->
        <div class="card mb-4">
          <div class="card-header">
            <h3 class="card-title">MBTI 人格类型</h3>
          </div>
          <div class="card-body">
            <div class="mbti-display">
              <span class="mbti-type">{{ mbtiDisplay }}</span>
              <span v-if="identity?.personality?.mbtiConfidence" class="confidence">
                置信度: {{ Math.round(identity.personality.mbtiConfidence * 100) }}%
              </span>
            </div>
            <div class="mbti-dimensions">
              <div class="dimension">
                <span class="dim-label">E</span>
                <div class="dim-bar">
                  <div class="dim-fill" :style="{ left: ((identity?.personality?.mbtiEI || 0) + 1) / 2 * 100 + '%' }"></div>
                </div>
                <span class="dim-label">I</span>
              </div>
              <div class="dimension">
                <span class="dim-label">N</span>
                <div class="dim-bar">
                  <div class="dim-fill" :style="{ left: ((identity?.personality?.mbtiSN || 0) + 1) / 2 * 100 + '%' }"></div>
                </div>
                <span class="dim-label">S</span>
              </div>
              <div class="dimension">
                <span class="dim-label">T</span>
                <div class="dim-bar">
                  <div class="dim-fill" :style="{ left: ((identity?.personality?.mbtiTF || 0) + 1) / 2 * 100 + '%' }"></div>
                </div>
                <span class="dim-label">F</span>
              </div>
              <div class="dimension">
                <span class="dim-label">J</span>
                <div class="dim-bar">
                  <div class="dim-fill" :style="{ left: ((identity?.personality?.mbtiJP || 0) + 1) / 2 * 100 + '%' }"></div>
                </div>
                <span class="dim-label">P</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Big Five -->
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">大五人格特质</h3>
          </div>
          <div class="card-body">
            <div class="trait-bars">
              <div v-for="trait in bigFiveScores" :key="trait.name" class="trait-item">
                <div class="trait-header">
                  <span class="trait-name">{{ trait.name }}</span>
                  <span class="trait-value">{{ Math.round(trait.value) }}%</span>
                </div>
                <div class="trait-bar">
                  <div class="trait-fill" :style="{ width: trait.value + '%', background: trait.color }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Values Tab -->
      <div v-if="activeTab === 'values'" class="profile-section">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">价值观维度</h3>
          </div>
          <div class="card-body">
            <div class="values-grid">
              <div v-for="value in valueScores" :key="value.name" class="value-item">
                <div class="value-circle" :style="{
                  background: `conic-gradient(#3B82F6 0% ${value.value}%, #E5E7EB ${value.value}% 100%)`
                }">
                  <div class="value-inner">
                    <span class="value-score">{{ Math.round(value.value) }}</span>
                  </div>
                </div>
                <span class="value-name">{{ value.name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Interests Tab -->
      <div v-if="activeTab === 'interests'" class="profile-section">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">兴趣爱好</h3>
            <button class="btn btn-outline btn-sm">添加兴趣</button>
          </div>
          <div class="card-body">
            <div v-if="identity?.interests?.length" class="interests-list">
              <div v-for="interest in identity.interests" :key="interest.id" class="interest-item">
                <div class="interest-level" :style="{ background: getInterestColor(interest.level) }">
                  {{ interest.level }}
                </div>
                <div class="interest-info">
                  <div class="interest-name">{{ interest.name }}</div>
                  <div class="interest-meta">
                    <span class="interest-category">{{ interest.category }}</span>
                    <span class="interest-desc">{{ interest.description }}</span>
                  </div>
                  <div v-if="interest.keywords?.length" class="interest-keywords">
                    <span v-for="kw in interest.keywords" :key="kw" class="keyword-tag">{{ kw }}</span>
                  </div>
                </div>
                <span class="interest-level-text">{{ getInterestLevel(interest.level) }}</span>
              </div>
            </div>
            <div v-else class="empty-state">
              <p class="text-muted">暂无兴趣爱好记录</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Memory Tab -->
      <div v-if="activeTab === 'memory'" class="profile-section">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">记忆库</h3>
            <button class="btn btn-primary btn-sm" @click="loadMemories">加载记忆</button>
          </div>
          <div class="card-body">
            <div v-if="memories.length" class="memories-list">
              <div v-for="memory in memories" :key="memory.id" class="memory-item">
                <div class="memory-type" :class="memory.type">
                  {{ memory.type }}
                </div>
                <div class="memory-content">
                  <div class="memory-key">{{ memory.key }}</div>
                  <div class="memory-text">{{ memory.content }}</div>
                  <div class="memory-meta">
                    <span>重要度: {{ memory.importance }}</span>
                    <span>访问: {{ memory.accessCount }}次</span>
                  </div>
                </div>
              </div>
            </div>
            <div v-else class="empty-state">
              <p class="text-muted">点击"加载记忆"查看记忆库</p>
            </div>
          </div>
        </div>
      </div>

      <!-- Preferences Tab -->
      <div v-if="activeTab === 'preferences'" class="profile-section">
        <div class="card">
          <div class="card-header">
            <h3 class="card-title">偏好设置</h3>
            <button class="btn btn-primary btn-sm" @click="loadPreferences">加载偏好</button>
          </div>
          <div class="card-body">
            <div v-if="preferences.length" class="preferences-list">
              <div v-for="pref in preferences" :key="pref.id" class="preference-item">
                <div class="pref-category">{{ pref.category }}</div>
                <div class="pref-content">
                  <span class="pref-key">{{ pref.key }}:</span>
                  <span class="pref-value">{{ JSON.stringify(pref.value) }}</span>
                </div>
                <div class="pref-confidence">
                  <div class="conf-bar">
                    <div class="conf-fill" :style="{ width: pref.confidence * 100 + '%' }"></div>
                  </div>
                  <span>{{ Math.round(pref.confidence * 100) }}%</span>
                </div>
              </div>
            </div>
            <div v-else class="empty-state">
              <p class="text-muted">点击"加载偏好"查看偏好设置</p>
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
.text-muted { color: var(--gray-500); }
.mt-1 { margin-top: 0.25rem; }
.mb-4 { margin-bottom: 1rem; }
.mb-6 { margin-bottom: 1.5rem; }

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

/* Tabs */
.tabs-container {
  display: flex;
  gap: 0.5rem;
  border-bottom: 1px solid var(--gray-200);
  padding-bottom: 0.5rem;
}

.tab-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: transparent;
  color: var(--gray-600);
  cursor: pointer;
  border-radius: 6px;
  font-size: 0.875rem;
  transition: all 0.2s;
}

.tab-btn:hover {
  background: var(--gray-100);
}

.tab-btn.active {
  background: var(--primary);
  color: white;
}

/* Profile Section */
.profile-section {
  max-width: 900px;
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1.5rem;
  background: white;
  border-radius: 12px;
  margin-bottom: 1.5rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.avatar-large {
  width: 80px;
  height: 80px;
  border-radius: 16px;
  background: linear-gradient(135deg, #6366F1 0%, #8B5CF6 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
}

.profile-name {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.profile-nickname {
  color: var(--gray-500);
  margin-top: 0.25rem;
}

.edit-btn {
  font-size: 0.75rem;
  padding: 0.25rem 0.5rem;
  background: var(--gray-100);
  border: none;
  border-radius: 4px;
  cursor: pointer;
  color: var(--gray-600);
}

.edit-btn:hover {
  background: var(--gray-200);
}

.edit-input {
  font-size: 1.25rem;
  padding: 0.25rem 0.5rem;
  border: 1px solid var(--primary);
  border-radius: 4px;
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

@media (max-width: 768px) {
  .info-grid { grid-template-columns: repeat(2, 1fr); }
}

.info-item {
  background: white;
  padding: 1rem;
  border-radius: 8px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.05);
}

.info-item label {
  display: block;
  font-size: 0.75rem;
  color: var(--gray-500);
  margin-bottom: 0.25rem;
}

.info-item span {
  font-weight: 500;
}

/* Card */
.card {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--gray-100);
}

.card-title {
  font-size: 1rem;
  font-weight: 600;
  margin: 0;
}

.card-body {
  padding: 1.25rem;
}

/* MBTI */
.mbti-display {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.mbti-type {
  font-size: 2rem;
  font-weight: 700;
  color: var(--primary);
}

.confidence {
  font-size: 0.875rem;
  color: var(--gray-500);
}

.mbti-dimensions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.dimension {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.dim-label {
  width: 24px;
  text-align: center;
  font-weight: 600;
  color: var(--gray-600);
}

.dim-bar {
  flex: 1;
  height: 8px;
  background: var(--gray-200);
  border-radius: 4px;
  position: relative;
}

.dim-fill {
  width: 12px;
  height: 12px;
  background: var(--primary);
  border-radius: 50%;
  position: absolute;
  top: -2px;
  transform: translateX(-50%);
  box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.2);
}

/* Trait Bars */
.trait-bars {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.trait-item {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.trait-header {
  display: flex;
  justify-content: space-between;
}

.trait-name {
  font-weight: 500;
  font-size: 0.875rem;
}

.trait-value {
  font-size: 0.875rem;
  color: var(--gray-500);
}

.trait-bar {
  height: 8px;
  background: var(--gray-200);
  border-radius: 4px;
  overflow: hidden;
}

.trait-fill {
  height: 100%;
  border-radius: 4px;
  transition: width 0.3s;
}

/* Values Grid */
.values-grid {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 1.5rem;
}

@media (max-width: 768px) {
  .values-grid { grid-template-columns: repeat(3, 1fr); }
}

.value-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
}

.value-circle {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.value-inner {
  width: 44px;
  height: 44px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.value-score {
  font-weight: 700;
  font-size: 0.875rem;
}

.value-name {
  font-size: 0.75rem;
  color: var(--gray-600);
}

/* Interests */
.interests-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.interest-item {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--gray-50);
  border-radius: 8px;
}

.interest-level {
  width: 28px;
  height: 28px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-weight: 700;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.interest-info {
  flex: 1;
}

.interest-name {
  font-weight: 500;
}

.interest-meta {
  display: flex;
  gap: 0.5rem;
  font-size: 0.75rem;
  color: var(--gray-500);
  margin-top: 0.25rem;
}

.interest-keywords {
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
  margin-top: 0.5rem;
}

.keyword-tag {
  font-size: 0.625rem;
  padding: 0.125rem 0.375rem;
  background: var(--gray-200);
  border-radius: 4px;
  color: var(--gray-600);
}

.interest-level-text {
  font-size: 0.75rem;
  color: var(--gray-500);
}

/* Memories */
.memories-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.memory-item {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--gray-50);
  border-radius: 8px;
}

.memory-type {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.625rem;
  font-weight: 600;
  text-transform: uppercase;
  background: var(--gray-200);
  color: var(--gray-600);
  height: fit-content;
}

.memory-type.preference { background: #DBEAFE; color: #1E40AF; }
.memory-type.fact { background: #D1FAE5; color: #065F46; }
.memory-type.event { background: #FEF3C7; color: #92400E; }

.memory-content {
  flex: 1;
}

.memory-key {
  font-weight: 500;
  font-size: 0.875rem;
}

.memory-text {
  font-size: 0.75rem;
  color: var(--gray-600);
  margin-top: 0.25rem;
}

.memory-meta {
  display: flex;
  gap: 0.75rem;
  font-size: 0.625rem;
  color: var(--gray-400);
  margin-top: 0.25rem;
}

/* Preferences */
.preferences-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.preference-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem;
  background: var(--gray-50);
  border-radius: 8px;
}

.pref-category {
  padding: 0.25rem 0.5rem;
  background: var(--gray-200);
  border-radius: 4px;
  font-size: 0.625rem;
  font-weight: 500;
  color: var(--gray-600);
}

.pref-content {
  flex: 1;
  font-size: 0.875rem;
}

.pref-key {
  font-weight: 500;
}

.pref-value {
  color: var(--gray-600);
  margin-left: 0.25rem;
}

.pref-confidence {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 100px;
}

.conf-bar {
  flex: 1;
  height: 4px;
  background: var(--gray-200);
  border-radius: 2px;
  overflow: hidden;
}

.conf-fill {
  height: 100%;
  background: var(--success);
}

/* Empty State */
.empty-state {
  text-align: center;
  padding: 2rem;
  color: var(--gray-400);
}
</style>