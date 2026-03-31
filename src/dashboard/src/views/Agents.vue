<script setup lang="ts">
import { ref, onMounted, inject, watch } from 'vue'
import { apiClient } from '../api/client'
import type { Agent } from '../types'

const agents = inject<any>('agents')
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
  if (!confirm('Are you sure you want to delete this agent?')) return
  try {
    await apiClient.deleteAgent(id)
    agents.value = agents.value.filter((a: Agent) => a.id !== id)
  } catch (e) {
    console.error('Failed to delete agent:', e)
    alert('Failed to delete agent')
  }
}

function viewAgent(agent: Agent) {
  selectedAgent.value = agent
}

function closeDetail() {
  selectedAgent.value = null
}

function getStatusClass(status: string): string {
  switch (status) {
    case 'online': return 'badge-success'
    case 'offline': return 'badge-danger'
    case 'busy': return 'badge-warning'
    default: return 'badge-gray'
  }
}

const filteredAgents = ref<Agent[]>([])

watch([agents, searchQuery, statusFilter], () => {
  let result = agents.value || []
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter((a: Agent) =>
      a.id.toLowerCase().includes(query) ||
      a.name.toLowerCase().includes(query) ||
      a.type.toLowerCase().includes(query)
    )
  }
  if (statusFilter.value) {
    result = result.filter((a: Agent) => a.status === statusFilter.value)
  }
  filteredAgents.value = result
}, { immediate: true })

onMounted(loadAgents)
</script>

<template>
  <div>
    <div class="flex items-center justify-between mb-4">
      <h2>Agents</h2>
      <button class="btn btn-outline" @click="loadAgents">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M23 4v6h-6M1 20v-6h6M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
        Refresh
      </button>
    </div>

    <!-- Filters -->
    <div class="card">
      <div class="flex gap-4">
        <div class="form-group" style="flex: 1;">
          <input
            type="text"
            class="form-input"
            placeholder="Search agents..."
            v-model="searchQuery"
          />
        </div>
        <div class="form-group">
          <select class="form-select" v-model="statusFilter">
            <option value="">All Status</option>
            <option value="online">Online</option>
            <option value="offline">Offline</option>
            <option value="busy">Busy</option>
            <option value="idle">Idle</option>
          </select>
        </div>
      </div>
    </div>

    <!-- Agent List -->
    <div class="card">
      <div class="table-container">
        <table v-if="filteredAgents.length">
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Type</th>
              <th>Status</th>
              <th>Capabilities</th>
              <th>Load</th>
              <th>Last Seen</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="agent in filteredAgents" :key="agent.id">
              <td><code>{{ agent.id?.slice(0, 12) }}</code></td>
              <td>{{ agent.name }}</td>
              <td>{{ agent.type }}</td>
              <td><span :class="['badge', getStatusClass(agent.status)]">{{ agent.status }}</span></td>
              <td>{{ agent.capabilities?.length || 0 }} skills</td>
              <td>
                <div class="flex items-center gap-2">
                  <div style="width: 60px; height: 6px; background: var(--gray-200); border-radius: 3px;">
                    <div
                      :style="{ width: agent.load + '%', background: agent.load > 80 ? 'var(--danger)' : 'var(--primary)', height: '100%', borderRadius: '3px' }"
                    ></div>
                  </div>
                  <span>{{ agent.load }}%</span>
                </div>
              </td>
              <td class="text-muted">{{ new Date(agent.lastSeen).toLocaleString() }}</td>
              <td>
                <div class="flex gap-2">
                  <button class="btn btn-outline btn-sm" @click="viewAgent(agent)">View</button>
                  <button class="btn btn-danger btn-sm" @click="deleteAgent(agent.id)">Delete</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
        <div v-else class="empty-state">
          <svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
            <circle cx="9" cy="7" r="4"></circle>
            <path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
            <path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
          </svg>
          <p>No agents found</p>
        </div>
      </div>
    </div>

    <!-- Agent Detail Modal -->
    <div v-if="selectedAgent" class="modal-overlay" @click="closeDetail">
      <div class="modal" @click.stop>
        <div class="modal-header">
          <h3>{{ selectedAgent.name }}</h3>
          <button class="btn btn-outline btn-sm" @click="closeDetail">Close</button>
        </div>
        <div class="modal-body">
          <div class="grid grid-2">
            <div>
              <div class="text-muted mb-1">Agent ID</div>
              <div><code>{{ selectedAgent.id }}</code></div>
            </div>
            <div>
              <div class="text-muted mb-1">Type</div>
              <div>{{ selectedAgent.type }}</div>
            </div>
            <div>
              <div class="text-muted mb-1">Status</div>
              <div><span :class="['badge', getStatusClass(selectedAgent.status)]">{{ selectedAgent.status }}</span></div>
            </div>
            <div>
              <div class="text-muted mb-1">Load</div>
              <div>{{ selectedAgent.load }}%</div>
            </div>
          </div>

          <div class="mt-4">
            <div class="text-muted mb-2">Capabilities</div>
            <div v-if="selectedAgent.capabilities?.length">
              <div v-for="cap in selectedAgent.capabilities" :key="cap.id" class="card" style="margin-bottom: 0.5rem; padding: 1rem;">
                <div class="flex items-center justify-between">
                  <div>
                    <strong>{{ cap.name }}</strong>
                    <div class="text-muted text-sm">{{ cap.id }}</div>
                  </div>
                  <div class="badge badge-info">{{ cap.operations?.length || 0 }} ops</div>
                </div>
                <div v-if="cap.operations?.length" class="mt-2">
                  <span v-for="op in cap.operations" :key="op" class="badge badge-gray" style="margin-right: 0.25rem;">{{ op }}</span>
                </div>
              </div>
            </div>
            <div v-else class="text-muted">No capabilities</div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal {
  background: white;
  border-radius: 8px;
  width: 600px;
  max-width: 90%;
  max-height: 80vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--gray-200);
}

.modal-body {
  padding: 1.5rem;
}
</style>