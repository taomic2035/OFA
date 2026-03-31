<script setup lang="ts">
import { ref, inject } from 'vue'
import { apiClient } from '../api/client'
import type { Agent } from '../types'

const agents = inject<any>('agents')

const messageType = ref<'direct' | 'broadcast'>('direct')
const sending = ref(false)
const messageHistory = ref<any[]>([])

const message = ref({
  from_agent: '',
  to_agent: '',
  action: '',
  payload: '{}'
})

const quickActions = [
  { action: 'ping', label: 'Ping 测试', payload: '{}' },
  { action: 'status', label: '获取状态', payload: '{}' },
  { action: 'sync', label: '同步数据', payload: '{}' },
  { action: 'health', label: '健康检查', payload: '{}' }
]

async function sendMessage() {
  sending.value = true
  try {
    const payload = JSON.parse(message.value.payload)

    let result
    if (messageType.value === 'broadcast') {
      result = await apiClient.broadcast({
        from_agent: message.value.from_agent,
        action: message.value.action,
        payload
      })
    } else {
      result = await apiClient.sendMessage({
        from_agent: message.value.from_agent,
        to_agent: message.value.to_agent,
        action: message.value.action,
        payload,
        require_ack: true
      })
    }

    messageHistory.value.unshift({
      id: Date.now(),
      type: messageType.value,
      from: message.value.from_agent || 'system',
      to: message.value.to_agent,
      action: message.value.action,
      success: result.success,
      timestamp: new Date().toISOString()
    })

    // Clear form
    message.value.action = ''
    message.value.payload = '{}'
  } catch (e) {
    console.error('Failed to send message:', e)
    alert('发送失败: ' + (e as Error).message)
  } finally {
    sending.value = false
  }
}

function applyQuickAction(action: string, payload: string) {
  message.value.action = action
  message.value.payload = payload
}

function formatTime(timestamp: string): string {
  return new Date(timestamp).toLocaleTimeString()
}

function clearHistory() {
  messageHistory.value = []
}
</script>

<template>
  <div>
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>消息中心</h1>
        <p class="text-muted">智能体间的消息通信</p>
      </div>
    </div>

    <div class="messages-layout">
      <!-- Send Panel -->
      <div class="send-panel">
        <div class="panel-header">
          <h3>发送消息</h3>
        </div>

        <div class="panel-body">
          <!-- Message Type -->
          <div class="form-group">
            <label class="form-label">消息类型</label>
            <div class="type-selector">
              <button
                :class="['type-btn', { active: messageType === 'direct' }]"
                @click="messageType = 'direct'"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <line x1="22" y1="2" x2="11" y2="13"/>
                  <polygon points="22 2 15 22 11 13 2 9 22 2"/>
                </svg>
                直接发送
              </button>
              <button
                :class="['type-btn', { active: messageType === 'broadcast' }]"
                @click="messageType = 'broadcast'"
              >
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="16" x2="12" y2="12"/>
                  <line x1="12" y1="8" x2="12.01" y2="8"/>
                </svg>
                广播
              </button>
            </div>
          </div>

          <!-- From / To -->
          <div class="form-row">
            <div class="form-group">
              <label class="form-label">发送者</label>
              <input type="text" class="form-input" v-model="message.from_agent" placeholder="智能体ID">
            </div>
            <div v-if="messageType === 'direct'" class="form-group">
              <label class="form-label">接收者</label>
              <select class="form-select" v-model="message.to_agent">
                <option value="">选择智能体...</option>
                <option v-for="agent in agents" :key="agent.id" :value="agent.id">
                  {{ agent.name }} ({{ agent.id?.slice(0, 8) }})
                </option>
              </select>
            </div>
          </div>

          <!-- Action -->
          <div class="form-group">
            <label class="form-label">动作</label>
            <input type="text" class="form-input" v-model="message.action" placeholder="如: execute, query, notify">
          </div>

          <!-- Payload -->
          <div class="form-group">
            <label class="form-label">数据 (JSON)</label>
            <textarea class="form-textarea" v-model="message.payload" rows="4" placeholder='{"key": "value"}'></textarea>
          </div>

          <!-- Send Button -->
          <button class="btn btn-primary btn-block" @click="sendMessage" :disabled="sending || !message.action">
            <svg v-if="sending" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spin">
              <line x1="12" y1="2" x2="12" y2="6"/>
              <line x1="12" y1="18" x2="12" y2="22"/>
              <line x1="4.93" y1="4.93" x2="7.76" y2="7.76"/>
              <line x1="16.24" y1="16.24" x2="19.07" y2="19.07"/>
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="22" y1="2" x2="11" y2="13"/>
              <polygon points="22 2 15 22 11 13 2 9 22 2"/>
            </svg>
            {{ sending ? '发送中...' : '发送消息' }}
          </button>
        </div>

        <!-- Quick Actions -->
        <div class="quick-actions">
          <div class="quick-actions-header">快捷操作</div>
          <div class="quick-actions-grid">
            <button
              v-for="qa in quickActions"
              :key="qa.action"
              class="quick-action-btn"
              @click="applyQuickAction(qa.action, qa.payload)"
            >
              {{ qa.label }}
            </button>
          </div>
        </div>
      </div>

      <!-- History Panel -->
      <div class="history-panel">
        <div class="panel-header">
          <h3>发送历史</h3>
          <button v-if="messageHistory.length" class="btn btn-outline btn-sm" @click="clearHistory">清空</button>
        </div>

        <div class="panel-body">
          <div v-if="messageHistory.length" class="history-list">
            <div v-for="msg in messageHistory" :key="msg.id" class="history-item">
              <div class="history-header">
                <span :class="['type-tag', msg.type]">{{ msg.type === 'broadcast' ? '广播' : '直接' }}</span>
                <span class="history-time">{{ formatTime(msg.timestamp) }}</span>
              </div>
              <div class="history-content">
                <div class="history-row">
                  <span class="history-label">动作:</span>
                  <span class="history-value">{{ msg.action }}</span>
                </div>
                <div class="history-row">
                  <span class="history-label">从:</span>
                  <span class="history-value">{{ msg.from?.slice(0, 8) || 'system' }}</span>
                </div>
                <div v-if="msg.to" class="history-row">
                  <span class="history-label">到:</span>
                  <span class="history-value">{{ msg.to?.slice(0, 8) }}</span>
                </div>
              </div>
              <div class="history-status">
                <span v-if="msg.success" class="status-success">✓ 发送成功</span>
                <span v-else class="status-failed">✗ 发送失败</span>
              </div>
            </div>
          </div>

          <div v-else class="empty-history">
            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
            <p>暂无消息记录</p>
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

.messages-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
}

@media (max-width: 1024px) {
  .messages-layout { grid-template-columns: 1fr; }
}

.send-panel, .history-panel {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
  overflow: hidden;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 1.25rem;
  background: var(--gray-50);
  border-bottom: 1px solid var(--gray-100);
}

.panel-header h3 { margin: 0; font-size: 1rem; }
.panel-body { padding: 1.25rem; }

.type-selector { display: flex; gap: 0.5rem; }

.type-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border: 2px solid var(--gray-200);
  background: white;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
}

.type-btn:hover { border-color: var(--gray-300); }
.type-btn.active { border-color: var(--primary); background: #EFF6FF; color: var(--primary); }

.form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }

.btn-block { width: 100%; margin-top: 1rem; }

.spin { animation: spin 1s linear infinite; }

.quick-actions {
  padding: 1rem 1.25rem;
  border-top: 1px solid var(--gray-100);
}

.quick-actions-header {
  font-size: 0.75rem;
  color: var(--gray-500);
  margin-bottom: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.quick-actions-grid { display: flex; flex-wrap: wrap; gap: 0.5rem; }

.quick-action-btn {
  padding: 0.5rem 1rem;
  background: var(--gray-100);
  border: none;
  border-radius: 6px;
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s;
}

.quick-action-btn:hover { background: var(--gray-200); }

.history-list { display: flex; flex-direction: column; gap: 0.75rem; max-height: 500px; overflow-y: auto; }

.history-item {
  padding: 0.875rem;
  background: var(--gray-50);
  border-radius: 8px;
  border-left: 3px solid var(--primary);
}

.history-item:has(.type-tag.broadcast) { border-left-color: var(--success); }

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.type-tag {
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.type-tag.direct { background: #DBEAFE; color: var(--primary); }
.type-tag.broadcast { background: #DCFCE7; color: var(--success); }

.history-time { font-size: 0.75rem; color: var(--gray-400); }

.history-content { display: flex; flex-direction: column; gap: 0.25rem; }

.history-row { display: flex; gap: 0.5rem; font-size: 0.8125rem; }
.history-label { color: var(--gray-500); }
.history-value { font-weight: 500; }

.history-status { margin-top: 0.5rem; font-size: 0.75rem; }
.status-success { color: var(--success); }
.status-failed { color: var(--danger); }

.empty-history {
  text-align: center;
  padding: 3rem;
  color: var(--gray-400);
}

.empty-history p { margin: 0.75rem 0 0 0; }
</style>