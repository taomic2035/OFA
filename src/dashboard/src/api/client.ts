import type {
  Agent,
  Task,
  Message,
  SystemInfo,
  SystemMetrics,
  Skill,
  ListAgentsResponse,
  ListTasksResponse,
  SubmitTaskRequest,
  SendMessageRequest,
  BroadcastRequest
} from '../types'

const API_BASE = '/api/v1'

class ApiClient {
  private baseUrl: string

  constructor(baseUrl: string = API_BASE) {
    this.baseUrl = baseUrl
  }

  private async request<T>(
    path: string,
    options: RequestInit = {}
  ): Promise<T> {
    const response = await fetch(`${this.baseUrl}${path}`, {
      headers: {
        'Content-Type': 'application/json',
        ...options.headers
      },
      ...options
    })

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Unknown error' }))
      throw new Error(error.error || `HTTP ${response.status}`)
    }

    return response.json()
  }

  // System
  async getSystemInfo(): Promise<SystemInfo> {
    return this.request<SystemInfo>('/system/info')
  }

  async getMetrics(): Promise<SystemMetrics> {
    return this.request<SystemMetrics>('/system/metrics')
  }

  // Agents
  async getAgents(params?: {
    type?: string
    status?: string
    page?: number
    pageSize?: number
  }): Promise<ListAgentsResponse> {
    const query = new URLSearchParams()
    if (params?.type) query.set('type', params.type)
    if (params?.status) query.set('status', params.status)
    if (params?.page) query.set('page', String(params.page))
    if (params?.pageSize) query.set('page_size', String(params.pageSize))

    const queryString = query.toString()
    return this.request<ListAgentsResponse>(`/agents${queryString ? '?' + queryString : ''}`)
  }

  async getAgent(id: string): Promise<Agent> {
    return this.request<Agent>(`/agents/${id}`)
  }

  async deleteAgent(id: string): Promise<{ success: boolean }> {
    return this.request<{ success: boolean }>(`/agents/${id}`, {
      method: 'DELETE'
    })
  }

  // Tasks
  async getTasks(params?: {
    status?: string
    agentId?: string
    page?: number
    pageSize?: number
  }): Promise<ListTasksResponse> {
    const query = new URLSearchParams()
    if (params?.status) query.set('status', params.status)
    if (params?.agentId) query.set('agent_id', params.agentId)
    if (params?.page) query.set('page', String(params.page))
    if (params?.pageSize) query.set('page_size', String(params.pageSize))

    const queryString = query.toString()
    return this.request<ListTasksResponse>(`/tasks${queryString ? '?' + queryString : ''}`)
  }

  async getTask(id: string): Promise<Task> {
    return this.request<Task>(`/tasks/${id}`)
  }

  async submitTask(task: SubmitTaskRequest): Promise<Task> {
    return this.request<Task>('/tasks', {
      method: 'POST',
      body: JSON.stringify(task)
    })
  }

  async cancelTask(id: string, reason?: string): Promise<{ success: boolean }> {
    return this.request<{ success: boolean }>(`/tasks/${id}/cancel`, {
      method: 'POST',
      body: JSON.stringify({ reason })
    })
  }

  // Skills
  async getSkills(category?: string): Promise<Skill[]> {
    const query = category ? `?category=${category}` : ''
    const response = await this.request<{ skills: Skill[] }>(`/skills${query}`)
    return response.skills || []
  }

  // Messages
  async sendMessage(msg: SendMessageRequest): Promise<{ success: boolean; messageId: string }> {
    return this.request('/messages', {
      method: 'POST',
      body: JSON.stringify(msg)
    })
  }

  async broadcast(msg: BroadcastRequest): Promise<{ success: boolean; count: number }> {
    return this.request('/messages/broadcast', {
      method: 'POST',
      body: JSON.stringify(msg)
    })
  }

  // Health
  async healthCheck(): Promise<{ status: string; version: string }> {
    return this.request('/health')
  }
}

// WebSocket Client
class WsClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 10
  private reconnectDelay = 3000
  private listeners: Map<string, Set<(data: any) => void>> = new Map()

  constructor(url: string = '/ws') {
    this.url = url.replace('http', 'ws')
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return

    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
        this.emit('connected', {})
      }

      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.emit(data.type || 'message', data)
        } catch (e) {
          console.error('Failed to parse WebSocket message:', e)
        }
      }

      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        this.emit('disconnected', {})
        this.attemptReconnect()
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        this.emit('error', { error })
      }
    } catch (e) {
      console.error('Failed to create WebSocket:', e)
      this.attemptReconnect()
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    console.log(`Reconnecting in ${this.reconnectDelay}ms (attempt ${this.reconnectAttempts})`)

    setTimeout(() => {
      this.connect()
    }, this.reconnectDelay)
  }

  on(event: string, callback: (data: any) => void): () => void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)

    // Return unsubscribe function
    return () => {
      this.listeners.get(event)?.delete(callback)
    }
  }

  private emit(event: string, data: any): void {
    this.listeners.get(event)?.forEach(cb => cb(data))
    this.listeners.get('*')?.forEach(cb => cb({ type: event, ...data }))
  }

  send(data: any): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }
}

export const apiClient = new ApiClient()
export const wsClient = new WsClient()