// OFA Dashboard Types

export interface Agent {
  id: string
  name: string
  type: string
  status: 'online' | 'offline' | 'busy' | 'idle'
  capabilities: Capability[]
  lastSeen: string
  load: number
  metadata?: Record<string, string>
}

export interface Capability {
  id: string
  name: string
  operations: string[]
  description?: string
  category?: string
}

export interface Task {
  id: string
  skillId: string
  operation: string
  input: any
  output?: any
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled'
  priority: number
  agentId?: string
  createdAt: string
  completedAt?: string
  error?: string
  duration?: number
}

export interface Message {
  id: string
  fromAgent: string
  toAgent: string
  action: string
  payload: any
  timestamp: string
  ttl?: number
  headers?: Record<string, string>
}

export interface SystemInfo {
  version: string
  uptime: number
  agentsTotal: number
  agentsOnline: number
  tasksTotal: number
  tasksCompleted: number
  tasksFailed: number
}

export interface SystemMetrics {
  agents: {
    total: number
    online: number
    offline: number
    busy: number
  }
  tasks: {
    total: number
    pending: number
    running: number
    completed: number
    failed: number
  }
  messages: {
    total: number
    delivered: number
    pending: number
  }
  performance: {
    avgTaskDuration: number
    avgResponseTime: number
    throughput: number
  }
}

export interface Skill {
  id: string
  name: string
  operations: string[]
  description?: string
  category?: string
}

// API Response Types

export interface ApiResponse<T = any> {
  success: boolean
  data?: T
  error?: string
}

export interface ListAgentsResponse {
  agents: Agent[]
  total: number
  page: number
  pageSize: number
}

export interface ListTasksResponse {
  tasks: Task[]
  total: number
  page: number
  pageSize: number
}

export interface SubmitTaskRequest {
  skill_id: string
  input: any
  target_agent?: string
  priority?: number
  timeout_ms?: number
  metadata?: Record<string, string>
}

export interface SendMessageRequest {
  from_agent: string
  to_agent: string
  action: string
  payload: any
  ttl?: number
  headers?: Record<string, string>
  require_ack?: boolean
  timeout_ms?: number
}

export interface BroadcastRequest {
  from_agent: string
  action: string
  payload: any
  ttl?: number
}