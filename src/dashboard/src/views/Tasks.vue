<script setup lang="ts">
import { ref, onMounted, inject, watch } from 'vue'
import { apiClient } from '../api/client'
import type { Task, Skill } from '../types'

const tasks = inject<any>('tasks')
const loading = ref(true)
const statusFilter = ref('')
const showSubmitForm = ref(false)
const skills = ref<Skill[]>([])

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
    alert('Failed to submit task: ' + (e as Error).message)
  }
}

async function cancelTask(id: string) {
  if (!confirm('Are you sure you want to cancel this task?')) return
  try {
    await apiClient.cancelTask(id)
    const task = tasks.value.find((t: Task) => t.id === id)
    if (task) task.status = 'cancelled'
  } catch (e) {
    console.error('Failed to cancel task:', e)
    alert('Failed to cancel task')
  }
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'completed': return 'badge-success'
    case 'failed':
    case 'cancelled': return 'badge-danger'
    case 'running': return 'badge-warning'
    case 'pending': return 'badge-info'
    default: return 'badge-gray'
  }
}

const filteredTasks = ref<Task[]>([])

watch([tasks, statusFilter], () => {
  let result = tasks.value || []
  if (statusFilter.value) {
    result = result.filter((t: Task) => t.status === statusFilter.value)
  }
  filteredTasks.value = result
}, { immediate: true })

onMounted(() => {
  loadTasks()
  loadSkills()
})
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2>Tasks</h2>
      <div class="flex gap-2">
        <button class="btn btn-outline" @click="loadTasks">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
          Refresh
        </button>
        <button class="btn btn-primary" @click="showSubmitForm = !showSubmitForm">
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"></line>
            <line x1="5" y1="12" x2="19" y2="12"></line>
          </svg>
          New Task
        </button>
      </div>
    </div>

    <!-- Submit Form -->
    <div v-if="showSubmitForm" class="card">
      <div class="card-header">
        <h3 class="card-title">Submit New Task</h3>
      </div>
      <div class="grid grid-2">
        <div class="form-group">
          <label class="form-label">Skill</label>
          <select class="form-select" v-model="newTask.skill_id">
            <option value="">Select a skill...</option>
            <option v-for="skill in skills" :key="skill.id" :value="skill.id">
              {{ skill.name }} ({{ skill.id }})
            </option>
          </select>
        </div>
        <div class="form-group">
          <label class="form-label">Target Agent (optional)</label>
          <input type="text" class="form-input" v-model="newTask.target_agent" placeholder="Leave empty for auto-assign" />
        </div>
        <div class="form-group">
          <label class="form-label">Input (JSON)</label>
          <textarea class="form-textarea" v-model="newTask.input" rows="4"></textarea>
        </div>
        <div class="form-group">
          <label class="form-label">Priority</label>
          <input type="number" class="form-input" v-model="newTask.priority" min="0" max="100" />
        </div>
      </div>
      <div class="flex gap-2 mt-4">
        <button class="btn btn-primary" @click="submitTask">Submit</button>
        <button class="btn btn-outline" @click="showSubmitForm = false">Cancel</button>
      </div>
    </div>

    <!-- Filters -->
    <div class="card">
      <div class="form-group">
        <select class="form-select" v-model="statusFilter" style="width: 200px;">
          <option value="">All Status</option>
          <option value="pending">Pending</option>
          <option value="running">Running</option>
          <option value="completed">Completed</option>
          <option value="failed">Failed</option>
          <option value="cancelled">Cancelled</option>
        </select>
      </div>
    </div>

    <!-- Task List -->
    <div class="card">
      <div class="table-container">
        <table v-if="filteredTasks.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>Skill</th>
              <th>Status</th>
              <th>Agent</th>
              <th>Priority</th>
              <th>Created</th>
              <th>Duration</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="task in filteredTasks" :key="task.id">
              <td><code>{{ task.id?.slice(0, 8) }}</code></td>
              <td>{{ task.skillId }}</td>
              <td><span :class="['badge', getStatusClass(task.status)]">{{ task.status }}</span></td>
              <td>{{ task.agentId?.slice(0, 8) || '-' }}</td>
              <td>{{ task.priority }}</td>
              <td class="text-muted">{{ new Date(task.createdAt).toLocaleString() }}</td>
              <td>{{ task.duration ? `${task.duration}ms` : '-' }}</td>
              <td>
                <button
                  v-if="task.status === 'pending' || task.status === 'running'"
                  class="btn btn-danger btn-sm"
                  @click="cancelTask(task.id)"
                >
                  Cancel
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">
          <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path>
            <polyline points="14 2 14 8 20 8"></polyline>
          </svg>
          <p>No tasks found</p>
        </div>
      </div>
    </div>
  </div>
</template>