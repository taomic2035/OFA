<script setup lang="ts">
import { ref, onMounted, inject } from 'vue'
import { apiClient } from '../api/client'
import type { Agent } from '../types'

const agents = inject<any>('agents')

const messageType = ref<'direct' | 'broadcast'>('direct')
const message = ref({
  from_agent: '',
  to_agent: '',
  action: '',
  payload: '{}'
})

const messageHistory = ref<any[]>([])
const sending = ref(false)

async function sendMessage() {
  sending.value = true
  try {
    const payload = JSON.parse(message.value.payload)

    if (messageType.value === 'broadcast') {
      const result = await apiClient.broadcast({
        from_agent: message.value.from_agent,
        action: message.value.action,
        payload
      })
      messageHistory.value.unshift({
        type: 'broadcast',
        ...message.value,
        result,
        timestamp: new Date().toISOString()
      })
    } else {
      const result = await apiClient.sendMessage({
        from_agent: message.value.from_agent,
        to_agent: message.value.to_agent,
        action: message.value.action,
        payload,
        require_ack: true
      })
      messageHistory.value.unshift({
        type: 'direct',
        ...message.value,
        result,
        timestamp: new Date().toISOString()
      })
    }

    // Clear form
    message.value.action = ''
    message.value.payload = '{}'
  } catch (e) {
    console.error('Failed to send message:', e)
    alert('Failed to send message: ' + (e as Error).message)
  } finally {
    sending.value = false
  }
}

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString()
}
</script>

<template>
  <div>
    <h2 class="mb-4">Messages</h2>

    <div class="grid grid-2">
      <!-- Send Message Form -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">Send Message</h3>
        </div>

        <div class="form-group">
          <label class="form-label">Message Type</label>
          <div class="flex gap-4">
            <label class="flex items-center gap-2">
              <input type="radio" v-model="messageType" value="direct" />
              Direct
            </label>
            <label class="flex items-center gap-2">
              <input type="radio" v-model="messageType" value="broadcast" />
              Broadcast
            </label>
          </div>
        </div>

        <div class="grid grid-2">
          <div class="form-group">
            <label class="form-label">From Agent</label>
            <input type="text" class="form-input" v-model="message.from_agent" placeholder="Source agent ID" />
          </div>

          <div class="form-group" v-if="messageType === 'direct'">
            <label class="form-label">To Agent</label>
            <select class="form-select" v-model="message.to_agent">
              <option value="">Select agent...</option>
              <option v-for="agent in agents" :key="agent.id" :value="agent.id">
                {{ agent.name }} ({{ agent.id?.slice(0, 8) }})
              </option>
            </select>
          </div>
        </div>

        <div class="form-group">
          <label class="form-label">Action</label>
          <input type="text" class="form-input" v-model="message.action" placeholder="e.g., execute, query, notify" />
        </div>

        <div class="form-group">
          <label class="form-label">Payload (JSON)</label>
          <textarea class="form-textarea" v-model="message.payload" rows="4" placeholder='{"key": "value"}'></textarea>
        </div>

        <button class="btn btn-primary" @click="sendMessage" :disabled="sending">
          {{ sending ? 'Sending...' : 'Send Message' }}
        </button>
      </div>

      <!-- Message History -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">Message History</h3>
        </div>

        <div v-if="messageHistory.length" class="message-list">
          <div v-for="(msg, index) in messageHistory" :key="index" class="message-item">
            <div class="flex items-center justify-between mb-2">
              <span :class="['badge', msg.type === 'broadcast' ? 'badge-info' : 'badge-gray']">{{ msg.type }}</span>
              <span class="text-muted text-sm">{{ formatTime(msg.timestamp) }}</span>
            </div>
            <div class="text-sm">
              <strong>From:</strong> {{ msg.from_agent?.slice(0, 8) || 'system' }}
              <span v-if="msg.to_agent"> → <strong>To:</strong> {{ msg.to_agent?.slice(0, 8) }}</span>
            </div>
            <div class="text-sm mt-1">
              <strong>Action:</strong> {{ msg.action }}
            </div>
            <div v-if="msg.result?.success" class="text-success text-sm mt-1">✓ Sent</div>
          </div>
        </div>

        <div v-else class="empty-state">
          <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
          </svg>
          <p>No messages sent yet</p>
        </div>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="card mt-4">
      <div class="card-header">
        <h3 class="card-title">Quick Actions</h3>
      </div>
      <div class="flex gap-2">
        <button class="btn btn-outline" @click="message.action = 'ping'; message.payload = '{}'">Ping All Agents</button>
        <button class="btn btn-outline" @click="message.action = 'status'; message.payload = '{}'">Request Status</button>
        <button class="btn btn-outline" @click="message.action = 'sync'; message.payload = '{}'">Sync Data</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.message-list {
  max-height: 400px;
  overflow-y: auto;
}

.message-item {
  padding: 1rem;
  border-bottom: 1px solid var(--gray-200);
}

.message-item:last-child {
  border-bottom: none;
}
</style>