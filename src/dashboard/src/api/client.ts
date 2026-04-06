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
  BroadcastRequest,
  PersonalIdentity,
  Memory,
  Preference,
  Decision,
  UserProfileResponse,
  ListMemoriesResponse,
  ListPreferencesResponse,
  ListDecisionsResponse
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

  // ===== User Profile APIs (v1.3.0) =====

  // Identity
  async getIdentity(userId: string): Promise<UserProfileResponse> {
    return this.request<UserProfileResponse>(`/users/${userId}/identity`)
  }

  async createIdentity(userId: string, identity: Partial<PersonalIdentity>): Promise<UserProfileResponse> {
    return this.request<UserProfileResponse>(`/users/${userId}/identity`, {
      method: 'POST',
      body: JSON.stringify(identity)
    })
  }

  async updateIdentity(userId: string, updates: Record<string, any>): Promise<UserProfileResponse> {
    return this.request<UserProfileResponse>(`/users/${userId}/identity`, {
      method: 'PATCH',
      body: JSON.stringify(updates)
    })
  }

  // Personality
  async getPersonality(userId: string): Promise<{ success: boolean; personality?: any; error?: string }> {
    return this.request(`/users/${userId}/personality`)
  }

  async updatePersonality(userId: string, updates: Record<string, any>): Promise<{ success: boolean; personality?: any; error?: string }> {
    return this.request(`/users/${userId}/personality`, {
      method: 'PATCH',
      body: JSON.stringify(updates)
    })
  }

  // Value System
  async getValueSystem(userId: string): Promise<{ success: boolean; valueSystem?: any; error?: string }> {
    return this.request(`/users/${userId}/values`)
  }

  async updateValueSystem(userId: string, updates: Record<string, any>): Promise<{ success: boolean; valueSystem?: any; error?: string }> {
    return this.request(`/users/${userId}/values`, {
      method: 'PATCH',
      body: JSON.stringify(updates)
    })
  }

  // Interests
  async getInterests(userId: string, category?: string): Promise<{ success: boolean; interests?: any[]; error?: string }> {
    const query = category ? `?category=${category}` : ''
    return this.request(`/users/${userId}/interests${query}`)
  }

  async addInterest(userId: string, interest: any): Promise<{ success: boolean; error?: string }> {
    return this.request(`/users/${userId}/interests`, {
      method: 'POST',
      body: JSON.stringify(interest)
    })
  }

  async removeInterest(userId: string, interestId: string): Promise<{ success: boolean; error?: string }> {
    return this.request(`/users/${userId}/interests/${interestId}`, {
      method: 'DELETE'
    })
  }

  // Memory
  async getMemories(userId: string, params?: {
    type?: string
    query?: string
    page?: number
    pageSize?: number
  }): Promise<ListMemoriesResponse> {
    const query = new URLSearchParams()
    if (params?.type) query.set('type', params.type)
    if (params?.query) query.set('query', params.query)
    if (params?.page) query.set('page', String(params.page))
    if (params?.pageSize) query.set('page_size', String(params.pageSize))

    const queryString = query.toString()
    return this.request<ListMemoriesResponse>(`/users/${userId}/memories${queryString ? '?' + queryString : ''}`)
  }

  async createMemory(userId: string, memory: Partial<Memory>): Promise<{ success: boolean; memory?: Memory; error?: string }> {
    return this.request(`/users/${userId}/memories`, {
      method: 'POST',
      body: JSON.stringify(memory)
    })
  }

  async deleteMemory(userId: string, memoryId: string): Promise<{ success: boolean; error?: string }> {
    return this.request(`/users/${userId}/memories/${memoryId}`, {
      method: 'DELETE'
    })
  }

  // Preferences
  async getPreferences(userId: string, category?: string): Promise<ListPreferencesResponse> {
    const query = category ? `?category=${category}` : ''
    return this.request<ListPreferencesResponse>(`/users/${userId}/preferences${query}`)
  }

  async setPreference(userId: string, category: string, key: string, value: any): Promise<{ success: boolean; preference?: Preference; error?: string }> {
    return this.request(`/users/${userId}/preferences`, {
      method: 'POST',
      body: JSON.stringify({ category, key, value })
    })
  }

  async deletePreference(userId: string, preferenceId: string): Promise<{ success: boolean; error?: string }> {
    return this.request(`/users/${userId}/preferences/${preferenceId}`, {
      method: 'DELETE'
    })
  }

  // Decisions
  async getDecisions(userId: string, params?: {
    context?: string
    page?: number
    pageSize?: number
  }): Promise<ListDecisionsResponse> {
    const query = new URLSearchParams()
    if (params?.context) query.set('context', params.context)
    if (params?.page) query.set('page', String(params.page))
    if (params?.pageSize) query.set('page_size', String(params.pageSize))

    const queryString = query.toString()
    return this.request<ListDecisionsResponse>(`/users/${userId}/decisions${queryString ? '?' + queryString : ''}`)
  }

  async getDecisionContext(userId: string): Promise<{ success: boolean; context?: any; error?: string }> {
    return this.request(`/users/${userId}/decision-context`)
  }

  // Voice Profile
  async getVoiceProfile(userId: string): Promise<{ success: boolean; voiceProfile?: any; error?: string }> {
    return this.request(`/users/${userId}/voice`)
  }

  async updateVoiceProfile(userId: string, profile: any): Promise<{ success: boolean; error?: string }> {
    return this.request(`/users/${userId}/voice`, {
      method: 'PATCH',
      body: JSON.stringify(profile)
    })
  }

  // Writing Style
  async getWritingStyle(userId: string): Promise<{ success: boolean; writingStyle?: any; error?: string }> {
    return this.request(`/users/${userId}/writing-style`)
  }

  async updateWritingStyle(userId: string, style: any): Promise<{ success: boolean; error?: string }> {
    return this.request(`/users/${userId}/writing-style`, {
      method: 'PATCH',
      body: JSON.stringify(style)
    })
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