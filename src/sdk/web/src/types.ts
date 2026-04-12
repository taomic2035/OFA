/**
 * OFA Web SDK - Type Definitions (v8.3.0)
 */

// Agent Status
export enum AgentStatus {
  Offline = 'offline',
  Online = 'online',
  Busy = 'busy',
  Error = 'error'
}

// Connection State
export enum ConnectionState {
  Disconnected = 'disconnected',
  Connecting = 'connecting',
  Connected = 'connected',
  Reconnecting = 'reconnecting',
  Error = 'error'
}

// Agent Mode
export enum AgentMode {
  Standalone = 'standalone',
  Sync = 'sync'
}

// Device Type
export enum DeviceType {
  Browser = 'browser',
  Desktop = 'desktop',
  Mobile = 'mobile',
  Unknown = 'unknown'
}

// WebSocket Message Types
export enum MessageType {
  Register = 'Register',
  RegisterAck = 'RegisterAck',
  Heartbeat = 'Heartbeat',
  StateUpdate = 'StateUpdate',
  TaskAssign = 'TaskAssign',
  TaskResult = 'TaskResult',
  SyncRequest = 'SyncRequest',
  SyncResponse = 'SyncResponse',
  BehaviorReport = 'BehaviorReport',
  EmotionUpdate = 'EmotionUpdate',
  Error = 'Error',
  AudioStream = 'audio_stream',
  AudioChunk = 'audio_chunk',
  AudioEnd = 'audio_end',
  TTSRequest = 'tts_request',
  Chat = 'chat',
  ChatStream = 'chat_stream'
}

// Scene Types
export enum SceneType {
  Running = 'running',
  Walking = 'walking',
  Driving = 'driving',
  Meeting = 'meeting',
  Sleeping = 'sleeping',
  Exercise = 'exercise',
  Work = 'work',
  Home = 'home',
  Travel = 'travel',
  HealthAlert = 'health_alert',
  Unknown = 'unknown'
}

// Audio Playback State
export enum AudioPlaybackState {
  Idle = 'idle',
  Playing = 'playing',
  Paused = 'paused',
  Stopped = 'stopped',
  Error = 'error'
}

// Agent Profile
export interface AgentProfile {
  agentId: string;
  deviceType: DeviceType;
  deviceName: string;
  browserInfo: string;
  capabilities: string[];
  identityId?: string;
  status: AgentStatus;
  lastHeartbeat: Date;
}

// WebSocket Message
export interface WebSocketMessage {
  type: MessageType;
  payload: Record<string, any>;
  timestamp: Date;
  messageId?: string;
}

// Personal Identity
export interface PersonalIdentity {
  id: string;
  name: string;
  nickname?: string;
  avatar?: string;
  personality: Personality;
  valueSystem: ValueSystem;
  interests: Interest[];
  voiceProfile: VoiceProfile;
  writingStyle: WritingStyle;
  version: number;
  updatedAt: Date;
}

// Big Five Personality Model
export interface Personality {
  openness: number;
  conscientiousness: number;
  extraversion: number;
  agreeableness: number;
  neuroticism: number;
}

// Value System
export interface ValueSystem {
  privacy: number;
  efficiency: number;
  health: number;
  family: number;
  career: number;
}

// Interest
export interface Interest {
  category: string;
  name: string;
  enthusiasm: number;
  keywords: string[];
}

// Voice Profile
export interface VoiceProfile {
  voiceType: string;
  pitch: number;
  speed: number;
  tone: string;
}

// Writing Style
export interface WritingStyle {
  formality: number;
  humor: number;
  emojiUsage: number;
}

// Behavior Observation
export interface BehaviorObservation {
  id: string;
  type: string;
  context: Record<string, any>;
  timestamp: Date;
}

// Scene State
export interface SceneState {
  id: string;
  type: SceneType;
  identityId: string;
  agentId: string;
  startTime: Date;
  endTime?: Date;
  confidence: number;
  context: Record<string, any>;
  actions: SceneAction[];
  active: boolean;
}

// Scene Action
export interface SceneAction {
  type: string;
  targetAgent?: string;
  payload: Record<string, any>;
  delay?: number;
}

// Chat Request
export interface ChatRequest {
  message: string;
  sessionId?: string;
  context?: Record<string, any>;
}

// Chat Response
export interface ChatResponse {
  response: string;
  sessionId: string;
  timestamp: Date;
  tokensUsed?: number;
}

// TTS Request
export interface TTSRequest {
  text: string;
  voice?: string;
  speed?: number;
}

// Audio Stream
export interface AudioStream {
  streamId: string;
  format: string;
  sampleRate: number;
  channels: number;
}

// Error Types
export enum OFAErrorCode {
  ConfigurationError = 'CONFIG_ERROR',
  ConnectionError = 'CONNECTION_ERROR',
  SyncError = 'SYNC_ERROR',
  IdentityError = 'IDENTITY_ERROR',
  AudioError = 'AUDIO_ERROR',
  ChatError = 'CHAT_ERROR'
}

export class OFAError extends Error {
  code: OFAErrorCode;
  
  constructor(code: OFAErrorCode, message: string) {
    super(message);
    this.code = code;
    this.name = 'OFAError';
  }
}
