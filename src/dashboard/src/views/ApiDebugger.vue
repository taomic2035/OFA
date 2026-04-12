<script setup lang="ts">
import { ref, computed } from 'vue'
import { apiClient } from '../api/client'

// API endpoints for testing
const apiEndpoints = [
  { path: '/health', method: 'GET', name: '健康检查', category: '系统' },
  { path: '/api/v1/status', method: 'GET', name: '系统状态', category: '系统' },
  { path: '/api/v1/system/info', method: 'GET', name: '系统信息', category: '系统' },
  { path: '/api/v1/system/metrics', method: 'GET', name: '系统指标', category: '系统' },
  { path: '/api/v1/agents', method: 'GET', name: '智能体列表', category: '智能体' },
  { path: '/api/v1/agents/{id}', method: 'GET', name: '智能体详情', category: '智能体', params: ['id'] },
  { path: '/api/v1/tasks', method: 'GET', name: '任务列表', category: '任务' },
  { path: '/api/v1/tasks', method: 'POST', name: '提交任务', category: '任务', body: true },
  { path: '/api/v1/tasks/{id}', method: 'GET', name: '任务详情', category: '任务', params: ['id'] },
  { path: '/api/v1/tasks/{id}/cancel', method: 'POST', name: '取消任务', category: '任务', params: ['id'] },
  { path: '/api/v1/skills', method: 'GET', name: '技能列表', category: '技能' },
  { path: '/api/v1/messages', method: 'POST', name: '发送消息', category: '消息', body: true },
  { path: '/api/v1/messages/broadcast', method: 'POST', name: '广播消息', category: '消息', body: true },
  { path: '/api/v1/scene', method: 'GET', name: '场景状态', category: '场景' },
  { path: '/api/v1/scene/active', method: 'GET', name: '活跃场景', category: '场景' },
  { path: '/api/v1/scene/history', method: 'GET', name: '场景历史', category: '场景' },
  { path: '/api/v1/scene/detect', method: 'POST', name: '场景检测', category: '场景', body: true },
  { path: '/api/v1/chat', method: 'POST', name: 'Chat对话', category: 'LLM', body: true },
  { path: '/api/v1/chat/stream', method: 'POST', name: '流式对话', category: 'LLM', body: true },
  { path: '/api/v1/chat/history', method: 'GET', name: '对话历史', category: 'LLM' },
  { path: '/api/v1/identity', method: 'POST', name: '创建身份', category: '身份', body: true },
  { path: '/api/v1/identity/{id}', method: 'GET', name: '获取身份', category: '身份', params: ['id'] },
  { path: '/api/v1/ws/connections', method: 'GET', name: 'WebSocket连接', category: 'WebSocket' },
]

// State
const selectedEndpoint = ref<any>(null)
const params = ref<Record<string, string>>({})
const requestBody = ref('')
const response = ref<any>(null)
const responseTime = ref(0)
const loading = ref(false)
const error = ref<string | null>(null)
const history = ref<any[]>([])
const wsMessages = ref<any[]>([])
const wsConnected = ref(false)
const wsInput = ref('')

// Computed
const categories = computed(() => {
  const cats = new Set(apiEndpoints.map(e => e.category))
  return Array.from(cats)
})

const filteredEndpoints = computed(() => {
  if (!selectedCategory.value) return apiEndpoints
  return apiEndpoints.filter(e => e.category === selectedCategory.value)
})

const selectedCategory = ref<string | null>(null)

// Methods
function selectEndpoint(endpoint: any) {
  selectedEndpoint.value = endpoint
  params.value = {}
  requestBody.value = getDefaultBody(endpoint)
  response.value = null
  error.value = null
}

function getDefaultBody(endpoint: any): string {
  if (!endpoint.body) return ''

  // Provide default bodies for common endpoints
  switch (endpoint.path) {
    case '/api/v1/tasks':
      return JSON.stringify({
        skill_id: 'search',
        input: { query: 'test query' },
        priority: 5
      }, null, 2)
    case '/api/v1/messages':
      return JSON.stringify({
        from_agent: 'test_agent_001',
        to_agent: 'test_agent_002',
        action: 'test',
        payload: { message: 'test message' }
      }, null, 2)
    case '/api/v1/messages/broadcast':
      return JSON.stringify({
        from_agent: 'test_agent_001',
        action: 'broadcast',
        payload: { message: 'broadcast message' }
      }, null, 2)
    case '/api/v1/scene/detect':
      return JSON.stringify({
        agent_id: 'test_agent_001',
        identity_id: 'test_identity_001',
        data: {
          activity_type: 'running',
          heart_rate: 145,
          duration: 1800
        }
      }, null, 2)
    case '/api/v1/chat':
      return JSON.stringify({
        message: 'Hello, this is a test message',
        session_id: 'test_session_001'
      }, null, 2)
    case '/api/v1/identity':
      return JSON.stringify({
        name: 'Test User',
        personality: {
          openness: 0.6,
          conscientiousness: 0.7,
          extraversion: 0.5,
          agreeableness: 0.8,
          neuroticism: 0.3
        }
      }, null, 2)
    default:
      return '{\n  \n}'
  }
}

async function executeRequest() {
  if (!selectedEndpoint.value) return

  loading.value = true
  error.value = null
  response.value = null
  const startTime = Date.now()

  try {
    // Build URL with params
    let url = selectedEndpoint.value.path
    for (const [key, value] of Object.entries(params.value)) {
      url = url.replace(`{${key}}`, value)
    }

    // Build query params
    const queryParams = new URLSearchParams()
    for (const [key, value] of Object.entries(params.value)) {
      if (!selectedEndpoint.value.path.includes(`{${key}}`)) {
        queryParams.set(key, value)
      }
    }
    if (queryParams.toString()) {
      url += '?' + queryParams.toString()
    }

    // Execute request
    const options: RequestInit = {
      method: selectedEndpoint.value.method,
      headers: { 'Content-Type': 'application/json' }
    }

    if (selectedEndpoint.value.method === 'POST' && requestBody.value) {
      options.body = requestBody.value
    }

    const res = await fetch(url, options)
    const endTime = Date.now()
    responseTime.value = endTime - startTime

    const contentType = res.headers.get('content-type') || ''

    if (contentType.includes('application/json')) {
      response.value = await res.json()
    } else if (contentType.includes('text/event-stream')) {
      // Handle SSE stream
      response.value = { status: 'Streaming...', type: 'SSE' }
      const reader = res.body?.getReader()
      if (reader) {
        const decoder = new TextDecoder()
        let result = ''
        while (true) {
          const { done, value } = await reader.read()
          if (done) break
          result += decoder.decode(value, { stream: true })
          response.value = { stream: result, type: 'SSE' }
        }
      }
    } else {
      response.value = { text: await res.text(), status: res.status }
    }

    // Add to history
    history.value.unshift({
      endpoint: selectedEndpoint.value,
      params: { ...params.value },
      body: requestBody.value,
      response: response.value,
      responseTime: responseTime.value,
      timestamp: new Date().toISOString(),
      status: res.status
    })
    if (history.value.length > 20) {
      history.value.pop()
    }

  } catch (e: any) {
    error.value = e.message
    responseTime.value = Date.now() - startTime
  } finally {
    loading.value = false
  }
}

function copyResponse() {
  if (response.value) {
    navigator.clipboard.writeText(JSON.stringify(response.value, null, 2))
  }
}

function loadFromHistory(item: any) {
  selectedEndpoint.value = item.endpoint
  params.value = { ...item.params }
  requestBody.value = item.body || ''
  response.value = item.response
  responseTime.value = item.responseTime
}

// WebSocket testing
let ws: WebSocket | null = null

function connectWs() {
  const wsUrl = 'ws://localhost:8080/ws'
  ws = new WebSocket(wsUrl)

  ws.onopen = () => {
    wsConnected.value = true
    wsMessages.value.push({ type: 'system', message: 'WebSocket connected', time: new Date().toLocaleTimeString() })
  }

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data)
      wsMessages.value.push({ type: 'received', data, time: new Date().toLocaleTimeString() })
    } catch (e) {
      wsMessages.value.push({ type: 'received', raw: event.data, time: new Date().toLocaleTimeString() })
    }
  }

  ws.onclose = () => {
    wsConnected.value = false
    wsMessages.value.push({ type: 'system', message: 'WebSocket disconnected', time: new Date().toLocaleTimeString() })
  }

  ws.onerror = (e) => {
    wsMessages.value.push({ type: 'error', message: 'WebSocket error', time: new Date().toLocaleTimeString() })
  }
}

function disconnectWs() {
  if (ws) {
    ws.close()
    ws = null
  }
}

function sendWsMessage() {
  if (ws && wsConnected.value && wsInput.value) {
    try {
      const data = JSON.parse(wsInput.value)
      ws.send(JSON.stringify(data))
      wsMessages.value.push({ type: 'sent', data, time: new Date().toLocaleTimeString() })
      wsInput.value = ''
    } catch (e) {
      ws.send(wsInput.value)
      wsMessages.value.push({ type: 'sent', raw: wsInput.value, time: new Date().toLocaleTimeString() })
    }
  }
}

function clearWsMessages() {
  wsMessages.value = []
}
</script>

<template>
  <div class="api-debugger">
    <!-- Header -->
    <div class="page-header">
      <div>
        <h1>API 调试工具</h1>
        <p class="text-muted">测试 Center REST API 和 WebSocket 接口</p>
      </div>
    </div>

    <div class="debugger-layout">
      <!-- Left: Endpoint List -->
      <div class="endpoint-panel">
        <div class="panel-header">
          <h3>API 端点</h3>
          <select v-model="selectedCategory" class="category-select">
            <option value="">全部</option>
            <option v-for="cat in categories" :key="cat" :value="cat">{{ cat }}</option>
          </select>
        </div>
        <div class="endpoint-list">
          <div v-for="endpoint in filteredEndpoints" :key="endpoint.path + endpoint.method"
               :class="['endpoint-item', { selected: selectedEndpoint === endpoint }]"
               @click="selectEndpoint(endpoint)">
            <span :class="['method-badge', endpoint.method.toLowerCase()]">{{ endpoint.method }}</span>
            <span class="endpoint-path">{{ endpoint.path }}</span>
            <span class="endpoint-name">{{ endpoint.name }}</span>
          </div>
        </div>
      </div>

      <!-- Middle: Request Builder -->
      <div class="request-panel">
        <div class="panel-header">
          <h3>请求构建</h3>
        </div>

        <div v-if="selectedEndpoint" class="request-content">
          <!-- Endpoint info -->
          <div class="endpoint-info">
            <span :class="['method-badge large', selectedEndpoint.method.toLowerCase()]">
              {{ selectedEndpoint.method }}
            </span>
            <span class="endpoint-url">{{ selectedEndpoint.path }}</span>
          </div>

          <!-- Params -->
          <div v-if="selectedEndpoint.params" class="params-section">
            <h4>参数</h4>
            <div v-for="param in selectedEndpoint.params" :key="param" class="param-input">
              <label>{{ param }}</label>
              <input v-model="params[param]" placeholder="输入值..." />
            </div>
          </div>

          <!-- Body -->
          <div v-if="selectedEndpoint.body" class="body-section">
            <h4>请求体 (JSON)</h4>
            <textarea v-model="requestBody" class="body-editor" placeholder="输入JSON请求体..."></textarea>
          </div>

          <!-- Execute -->
          <div class="execute-section">
            <button class="btn btn-primary" @click="executeRequest" :disabled="loading">
              {{ loading ? '执行中...' : '执行请求' }}
            </button>
            <span v-if="responseTime" class="response-time">{{ responseTime }}ms</span>
          </div>
        </div>

        <div v-else class="empty-state">
          <p>选择一个端点开始调试</p>
        </div>
      </div>

      <!-- Right: Response Panel -->
      <div class="response-panel">
        <div class="panel-header">
          <h3>响应结果</h3>
          <button v-if="response" class="btn btn-outline btn-sm" @click="copyResponse">复制</button>
        </div>

        <div v-if="loading" class="loading-state">
          <div class="loading-spinner"></div>
          <p>请求中...</p>
        </div>

        <div v-else-if="error" class="error-state">
          <div class="error-icon">❌</div>
          <p class="error-message">{{ error }}</p>
        </div>

        <div v-else-if="response" class="response-content">
          <pre class="response-json">{{ JSON.stringify(response, null, 2) }}</pre>
        </div>

        <div v-else class="empty-state">
          <p>等待响应...</p>
        </div>
      </div>
    </div>

    <!-- WebSocket Test -->
    <div class="ws-panel">
      <div class="panel-header">
        <h3>WebSocket 测试</h3>
        <div class="ws-controls">
          <button v-if="!wsConnected" class="btn btn-outline" @click="connectWs">连接</button>
          <button v-else class="btn btn-danger" @click="disconnectWs">断开</button>
          <button class="btn btn-outline" @click="clearWsMessages">清空消息</button>
        </div>
      </div>

      <div class="ws-content">
        <div class="ws-status">
          <span :class="['ws-status-dot', wsConnected ? 'connected' : 'disconnected']"></span>
          {{ wsConnected ? '已连接' : '未连接' }}
        </div>

        <div class="ws-messages">
          <div v-for="(msg, i) in wsMessages" :key="i" :class="['ws-message', msg.type]">
            <span class="msg-time">{{ msg.time }}</span>
            <span class="msg-type">{{ msg.type }}</span>
            <pre class="msg-content">{{ msg.data ? JSON.stringify(msg.data, null, 2) : msg.message || msg.raw }}</pre>
          </div>
        </div>

        <div class="ws-input-section">
          <textarea v-model="wsInput" class="ws-input" placeholder="输入JSON消息..."></textarea>
          <button class="btn btn-primary" @click="sendWsMessage" :disabled="!wsConnected">发送</button>
        </div>
      </div>
    </div>

    <!-- History -->
    <div class="history-panel">
      <div class="panel-header">
        <h3>请求历史</h3>
      </div>
      <div v-if="history.length" class="history-list">
        <div v-for="(item, i) in history" :key="i" class="history-item" @click="loadFromHistory(item)">
          <span :class="['method-badge', item.endpoint.method.toLowerCase()]">{{ item.endpoint.method }}</span>
          <span class="history-path">{{ item.endpoint.path }}</span>
          <span class="history-time">{{ item.responseTime }}ms</span>
          <span class="history-status">{{ item.status }}</span>
        </div>
      </div>
      <div v-else class="empty-state">
        <p>暂无请求历史</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.api-debugger { padding: 0; }

.page-header { margin-bottom: 1.5rem; }
.page-header h1 { margin: 0; font-size: 1.5rem; }
.page-header p { margin: 0.25rem 0 0 0; }

.debugger-layout {
  display: grid;
  grid-template-columns: 250px 1fr 1fr;
  gap: 1rem;
  margin-bottom: 1rem;
}

@media (max-width: 1200px) {
  .debugger-layout { grid-template-columns: 1fr 1fr; }
  .endpoint-panel { grid-column: span 2; }
}

@media (max-width: 768px) {
  .debugger-layout { grid-template-columns: 1fr; }
}

.panel { background: white; border-radius: 12px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid var(--gray-100);
}
.panel-header h3 { margin: 0; font-size: 1rem; }

.category-select {
  padding: 0.5rem;
  border: 1px solid var(--gray-200);
  border-radius: 6px;
  font-size: 0.875rem;
}

/* Endpoint Panel */
.endpoint-panel { background: white; border-radius: 12px; }

.endpoint-list { padding: 0.5rem; max-height: 400px; overflow-y: auto; }

.endpoint-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}
.endpoint-item:hover { background: var(--gray-50); }
.endpoint-item.selected { background: var(--primary-light); }

.method-badge {
  padding: 0.25rem 0.5rem;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 600;
}
.method-badge.get { background: #DBEAFE; color: #3B82F6; }
.method-badge.post { background: #DCFCE7; color: #22C55E; }
.method-badge.put { background: #FEF3C7; color: #F59E0B; }
.method-badge.delete { background: #FEE2E2; color: #EF4444; }
.method-badge.large { padding: 0.5rem 1rem; font-size: 0.875rem; }

.endpoint-path { font-size: 0.875rem; color: var(--gray-600); }
.endpoint-name { font-size: 0.875rem; }

/* Request Panel */
.request-panel { background: white; border-radius: 12px; }

.request-content { padding: 1rem; }

.endpoint-info { display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem; }
.endpoint-url { font-size: 1rem; font-weight: 500; }

.params-section, .body-section { margin-bottom: 1rem; }
.params-section h4, .body-section h4 { font-size: 0.875rem; margin: 0 0 0.5rem 0; color: var(--gray-600); }

.param-input { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.5rem; }
.param-input label { min-width: 80px; font-size: 0.875rem; }
.param-input input {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid var(--gray-200);
  border-radius: 6px;
}

.body-editor {
  width: 100%;
  min-height: 150px;
  padding: 0.75rem;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-family: monospace;
  font-size: 0.875rem;
  resize: vertical;
}

.execute-section { display: flex; align-items: center; gap: 1rem; }
.response-time { font-size: 0.875rem; color: var(--gray-500); }

/* Response Panel */
.response-panel { background: white; border-radius: 12px; }

.response-content { padding: 1rem; max-height: 400px; overflow-y: auto; }
.response-json {
  font-family: monospace;
  font-size: 0.875rem;
  white-space: pre-wrap;
  background: var(--gray-50);
  padding: 1rem;
  border-radius: 8px;
  margin: 0;
}

.loading-state, .error-state, .empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  text-align: center;
}

.loading-spinner {
  width: 30px;
  height: 30px;
  border: 3px solid var(--gray-200);
  border-top-color: var(--primary);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.error-icon { font-size: 2rem; margin-bottom: 0.5rem; }
.error-message { color: var(--danger); }

/* WebSocket Panel */
.ws-panel {
  background: white;
  border-radius: 12px;
  margin-bottom: 1rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.ws-controls { display: flex; gap: 0.5rem; }

.ws-content { padding: 1rem; }

.ws-status {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.ws-status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}
.ws-status-dot.connected { background: var(--success); }
.ws-status-dot.disconnected { background: var(--danger); }

.ws-messages {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  padding: 0.5rem;
  margin-bottom: 1rem;
}

.ws-message {
  display: flex;
  align-items: flex-start;
  gap: 0.5rem;
  padding: 0.25rem 0;
  border-bottom: 1px solid var(--gray-100);
}

.ws-message.sent { background: var(--gray-50); }
.ws-message.received { background: #E0F2FE; }
.ws-message.system { background: #F0FDF4; }
.ws-message.error { background: #FEE2E2; }

.msg-time { font-size: 0.75rem; color: var(--gray-500); min-width: 80px; }
.msg-type {
  font-size: 0.75rem;
  padding: 0.125rem 0.5rem;
  border-radius: 4px;
  background: var(--gray-200);
  min-width: 60px;
}
.msg-content {
  font-family: monospace;
  font-size: 0.75rem;
  margin: 0;
  white-space: pre-wrap;
}

.ws-input-section { display: flex; gap: 0.5rem; }
.ws-input {
  flex: 1;
  padding: 0.5rem;
  border: 1px solid var(--gray-200);
  border-radius: 8px;
  font-family: monospace;
  font-size: 0.875rem;
  min-height: 60px;
}

/* History Panel */
.history-panel {
  background: white;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.history-list { padding: 0.5rem; max-height: 200px; overflow-y: auto; }

.history-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
}
.history-item:hover { background: var(--gray-50); }

.history-path { flex: 1; font-size: 0.875rem; }
.history-time { font-size: 0.75rem; color: var(--gray-500); }
.history-status { font-size: 0.75rem; color: var(--success); }

@keyframes spin { to { transform: rotate(360deg); } }
</style>