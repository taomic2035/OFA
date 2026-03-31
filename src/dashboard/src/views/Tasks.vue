<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { apiClient } from '../api/client'
import type { Task, Skill } from '../types'

const tasks = ref<Task[]>([])
const skills = ref<Skill[]>([])
const loading = ref(true)
const statusFilter = ref('')
const showSubmitForm = ref(false)

const newTask = ref({
  skill_id: '',
  input: '{}',
  target_agent: '',
  priority: 0
})

async function loadTasks() {
  loading.value = true
  try {
    const response = await apiClient.getTasks({
      status: statusFilter.value || undefined
    })
    tasks.value = response.tasks || []
  } catch (e) {
    console.error('Failed to load tasks:', e)
  } finally {
    loading.value = false
  }
}

async function loadSkills() {
  try {
    skills.value = await apiClient.getSkills()
  } catch (e) {
    console.error('Failed to load skills:', e)
  }
}

async function submitTask() {
  try {
    const task = await apiClient.submitTask({
      skill_id: newTask.value.skill_id,
      input: JSON.parse(newTask.value.input),
      target_agent: newTask.value.target_agent || undefined,
      priority: newTask.value.priority || undefined
    })
    tasks.value = [task, ...tasks.value]
    showSubmitForm.value = false
    newTask.value = { skill_id: '', input: '{}', target_agent: '', priority: 0 }
  } catch (e) {
    console.error('Failed to submit task:', e)
    alert('提交失败: ' + (e as Error).message)
  }
}

async function cancelTask(id: string) {
  if (!confirm('确定要取消此任务吗？')) return
  try {
    await apiClient.cancelTask(id)
    const task = tasks.value.find(t => t.id === id)
    if (task) task.status = 'cancelled'
  } catch (e) {
    console.error('Failed to cancel task:', e)
    alert('取消失败')
  }
}

const filteredTasks = computed(() => {
  if (!statusFilter.value) return tasks.value
  return tasks.value.filter(t => t.status === statusFilter.value)
})

const taskStats = computed(() => ({
  total: tasks.value.length,
  completed: tasks.value.filter(t => t.status === 'completed').length,
  running: tasks.value.filter(t => t.status === 'running').length,
  failed: tasks.value.filter(t => t.status === 'failed').length
}))

function getStatusClass(status: string): string {
  switch (status) {
    case 'completed': return 'status-completed'
    case 'failed':
    case 'cancelled': return 'status-failed'
    case 'running': return 'status-running'
    default: return 'status-pending'
  }
}

function getStatusBadge(status: string): string {
  switch (status) {
    case 'completed': return 'badge-success'
    case 'failed':
    case 'cancelled': return 'badge-danger'
    case 'running': return 'badge-warning'
    default: return 'badge-info'
  }
}

function formatTime(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  loadTasks()
  loadSkills()
})
</script>

<template>
  <div>
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>任务管理</h1>
        <p class="text-muted">管理和监控所有任务</p>
      </div>
      <div class="header-actions">
        <button class="btn btn-outline" @click="loadTasks" :disabled="loading">刷新</button>
        <button class="btn btn-primary" @click="showSubmitForm = !showSubmitForm">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新建任务
        </button>
      </div>
    </div>

    <!-- Submit Form -->
    <div v-if="showSubmitForm" class="submit-panel">
      <div class="panel-header">
        <h3>提交新任务</h3>
        <button class="btn btn-outline btn-sm" @click="showSubmitForm = false">取消</button>
      </div>
      <div class="panel-body">
        <div class="form-grid">
          <div class="form-group">
            <label class="form-label">技能</label>
            <select class="form-select" v-model="newTask.skill_id">
              <option value="">选择技能...</option>
              <option v-for="skill in skills" :key="skill.id" :value="skill.id">
                {{ skill.name }} ({{ skill.id }})
              </option>
            </select>
          </div>
          <div class="form-group">
            <label class="form-label">目标智能体 (可选)</label>
            <input type="text" class="form-input" v-model="newTask.target_agent" placeholder="留空自动分配">
          </div>
          <div class="form-group">
            <label class="form-label">优先级</label>
            <input type="number" class="form-input" v-model="newTask.priority" min="0" max="100">
          </div>
        </div>
        <div class="form-group">
          <label class="form-label">输入参数 (JSON)</label>
          <textarea class="form-textarea" v-model="newTask.input" rows="3" placeholder='{"key": "value"}'></textarea>
        </div>
        <div class="form-actions">
          <button class="btn btn-primary" @click="submitTask">提交任务</button>
        </div>
      </div>
    </div>

    <!-- Stats -->
    <div class="stats-bar">
      <div class="stat-item">
        <span class="stat-num">{{ taskStats.total }}</span>
        <span class="stat-text">总任务</span>
      </div>
      <div class="stat-item">
        <span class="stat-num text-success">{{ taskStats.completed }}</span>
        <span class="stat-text">已完成</span>
      </div>
      <div class="stat-item">
        <span class="stat-num text-warning">{{ taskStats.running }}</span>
        <span class="stat-text">运行中</span>
      </div>
      <div class="stat-item">
        <span class="stat-num text-danger">{{ taskStats.failed }}</span>
        <span class="stat-text">失败</span>
      </div>
    </div>

    <!-- Filters -->
    <div class="filter-bar">
      <div class="filter-tabs">
        <button :class="['filter-tab', { active: statusFilter === '' }]" @click="statusFilter = ''">全部</button>
        <button :class="['filter-tab', { active: statusFilter === 'pending' }]" @click="statusFilter = 'pending'">等待</button>
        <button :class="['filter-tab', { active: statusFilter === 'running' }]" @click="statusFilter = 'running'">运行</button>
        <button :class="['filter-tab', { active: statusFilter === 'completed' }]" @click="statusFilter = 'completed'">完成</button>
        <button :class="['filter-tab', { active: statusFilter === 'failed' }]" @click="statusFilter = 'failed'">失败</button>
      </div>
    </div>

    <!-- Task List -->
    <div v-if="loading" class="loading-state">
      <div class="loading-spinner"></div>
    </div>

    <div v-else-if="filteredTasks.length === 0" class="empty-state">
      <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
        <polyline points="14 2 14 8 20 8"/>
      </svg>
      <h3>暂无任务</h3>
      <p class="text-muted">点击"新建任务"创建第一个任务</p>
    </div>

    <div v-else class="task-list">
      <div v-for="task in filteredTasks" :key="task.id" class="task-item">
        <div :class="['task-status-icon', getStatusClass(task.status)]">
          <svg v-if="task.status === 'completed'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          <svg v-else-if="task.status === 'failed' || task.status === 'cancelled'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="15" y1="9" x2="9" y2="15"/>
            <line x1="9" y1="9" x2="15" y2="15"/>
          </svg>
          <svg v-else-if="task.status === 'running'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin">
            <line x1="12" y1="2" x2="12" y2="6"/>
            <line x1="12" y1="18" x2="12" y2="22"/>
            <line x1="4.93" y1="4.93" x2="7.76" y2="7.76"/>
            <line x1="16.24" y1="16.24" x2="19.07" y2="19.07"/>
            <line x1="2" y1="12" x2="6" y2="12"/>
            <line x1="18" y1="12" x2="22" y2="12"/>
            <line x1="4.93" y1="19.07" x2="7.76" y2="16.24"/>
            <line x1="16.24" y1="7.76" x2="19.07" y2="4.93"/>
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
        </div>

        <div class="task-content">
          <div class="task-header">
            <span class="task-skill">{{ task.skillId }}</span>
            <span :class="['status-badge', getStatusBadge(task.status)]">{{ task.status }}</span>
          </div>
          <div class="task-meta">
            <span class="meta-item">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                <line x1="16" y1="2" x2="16" y2="6"/>
                <line x1="8" y1="2" x2="8" y2="6"/>
                <line x1="3" y1="10" x2="21" y2="10"/>
              </svg>
              {{ formatTime(task.createdAt) }}
            </span>
            <span v-if="task.duration" class="meta-item">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/>
                <polyline points="12 6 12 12 16 14"/>
              </svg>
              {{ task.duration }}ms
            </span>
            <span v-if="task.agentId" class="meta-item">
              <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                <circle cx="12" cy="7" r="4"/>
              </svg>
              {{ task.agentId?.slice(0, 8) }}
            </span>
          </div>
        </div>

        <div class="task-actions">
          <button
            v-if="task.status === 'pending' || task.status === 'running'"
            class="btn btn-danger btn-sm"
            @click="cancelTask(task.id)"
          >
            取消
          </button>
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

.submit-panel {
  background: white;
  border-radius: 12px;
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
  margin-bottom: 1.5rem;
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  background: var(--gray-50);
  border-bottom: 1px solid var(--gray-100);
}

.panel-header h3 { margin: 0; font-size: 1rem; }
.panel-body { padding: 1.5rem; }

.form-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 1rem; }
.form-actions { margin-top: 1rem; }

.stats-bar {
  display: flex;
  gap: 2rem;
  padding: 1rem 1.5rem;
  background: white;
  border-radius: 12px;
  margin-bottom: 1rem;
}

.stat-item { display: flex; flex-direction: column; }
.stat-num { font-size: 1.5rem; font-weight: 700; }
.stat-text { font-size: 0.75rem; color: var(--gray-500); }

.filter-bar { margin-bottom: 1rem; }

.filter-tabs {
  display: flex;
  background: white;
  border-radius: 8px;
  padding: 4px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
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

.filter-tab:hover { background: var(--gray-100); }
.filter-tab.active { background: var(--primary); color: white; }

.task-list { display: flex; flex-direction: column; gap: 0.75rem; }

.task-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.task-status-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.task-status-icon.status-completed { background: #DCFCE7; color: var(--success); }
.task-status-icon.status-failed { background: #FEE2E2; color: var(--danger); }
.task-status-icon.status-running { background: #FEF3C7; color: var(--warning); }
.task-status-icon.status-pending { background: #DBEAFE; color: var(--primary); }

.spin { animation: spin 1s linear infinite; }
@keyframes spin { to { transform: rotate(360deg); } }

.task-content { flex: 1; min-width: 0; }

.task-header { display: flex; align-items: center; gap: 0.75rem; margin-bottom: 0.375rem; }
.task-skill { font-weight: 600; }

.status-badge {
  padding: 0.125rem 0.625rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 500;
}

.task-meta { display: flex; flex-wrap: wrap; gap: 1rem; }
.meta-item { display: flex; align-items: center; gap: 0.375rem; font-size: 0.8125rem; color: var(--gray-500); }

.task-actions { flex-shrink: 0; }

.loading-state { display: flex; justify-content: center; padding: 4rem; }
.loading-spinner { width: 40px; height: 40px; border: 3px solid var(--gray-200); border-top-color: var(--primary); border-radius: 50%; animation: spin 1s linear infinite; }

.empty-state { text-align: center; padding: 4rem; }
.empty-state h3 { margin: 1rem 0 0.5rem 0; }

@media (max-width: 768px) {
  .form-grid { grid-template-columns: 1fr; }
  .stats-bar { flex-wrap: wrap; }
}
</style>