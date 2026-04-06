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

// ===== User Profile Types (v1.3.0) =====

export interface PersonalIdentity {
  id: string
  name: string
  nickname?: string
  avatar?: string
  birthday?: string
  gender?: string
  location?: string
  occupation?: string
  languages?: string[]
  timezone?: string
  personality?: Personality
  valueSystem?: ValueSystem
  interests?: Interest[]
  voiceProfile?: VoiceProfile
  writingStyle?: WritingStyle
  createdAt: string
  updatedAt: string
}

export interface Personality {
  // MBTI
  mbtiType?: string
  mbtiEI?: number      // -1.0 (I) to 1.0 (E)
  mbtiSN?: number      // -1.0 (S) to 1.0 (N)
  mbtiTF?: number      // -1.0 (T) to 1.0 (F)
  mbtiJP?: number      // -1.0 (J) to 1.0 (P)
  mbtiConfidence?: number

  // Big Five
  openness?: number
  conscientiousness?: number
  extraversion?: number
  agreeableness?: number
  neuroticism?: number

  // Stability
  stabilityScore?: number
  observedCount?: number

  // Speaking
  speakingTone?: string
  responseLength?: string
  emojiUsage?: string

  // Tags
  customTraits?: Record<string, number>
  summary?: string
  tags?: string[]
}

export interface ValueSystem {
  privacy?: number
  efficiency?: number
  health?: number
  family?: number
  career?: number
  entertainment?: number
  learning?: number
  social?: number
  finance?: number
  environment?: number
  riskTolerance?: number
  impulsiveness?: number
  patience?: number
  customValues?: Record<string, number>
  summary?: string
}

export interface Interest {
  id: string
  category: string
  name: string
  level: number    // 1-5
  keywords?: string[]
  description?: string
  since?: string
  lastActive?: string
}

export interface VoiceProfile {
  id: string
  voiceType?: string
  presetVoiceId?: string
  cloneReferenceId?: string
  pitch?: number
  speed?: number
  volume?: number
  tone?: string
  accent?: string
  emotionLevel?: number
  pausePattern?: string
  emphasisStyle?: string
  createdAt?: string
  updatedAt?: string
}

export interface WritingStyle {
  formality?: number
  verbosity?: number
  humor?: number
  technicality?: number
  useEmoji?: boolean
  useGifs?: boolean
  useMarkdown?: boolean
  signaturePhrase?: string
  frequentWords?: string[]
  avoidWords?: string[]
  preferredGreeting?: string
  preferredClosing?: string
}

export interface Memory {
  id: string
  type: 'preference' | 'fact' | 'event' | 'context' | 'skill' | 'relation' | 'feedback'
  key: string
  content: string
  importance: number
  accessCount: number
  lastAccessedAt: string
  expiresAt?: string
  metadata?: Record<string, any>
  createdAt: string
  updatedAt: string
}

export interface Preference {
  id: string
  category: string
  key: string
  value: any
  confidence: number
  source: string
  validUntil?: string
  createdAt: string
  updatedAt: string
}

export interface Decision {
  id: string
  context: string
  options: DecisionOption[]
  selectedOption: string
  reasoning?: string
  confidence: number
  factors: DecisionFactor[]
  outcome?: string
  userFeedback?: string
  createdAt: string
}

export interface DecisionOption {
  id: string
  name: string
  description?: string
  score: number
}

export interface DecisionFactor {
  name: string
  weight: number
  value: number
  description?: string
}

// API Response Types

export interface UserProfileResponse {
  success: boolean
  identity?: PersonalIdentity
  error?: string
}

export interface ListMemoriesResponse {
  memories: Memory[]
  total: number
}

export interface ListPreferencesResponse {
  preferences: Preference[]
  total: number
}

export interface ListDecisionsResponse {
  decisions: Decision[]
  total: number
}