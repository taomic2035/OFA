/**
 * Agent Configuration
 * Sprint 29: Node.js Agent SDK
 */

export interface AgentConfig {
  agentId: string;
  name: string;
  centerUrl: string;
  connectionType: 'grpc' | 'http' | 'websocket';
  heartbeatInterval: number;
  reconnectInterval: number;
  maxReconnectAttempts: number;
  skills: string[];
  metadata: Record<string, unknown>;
  tlsEnabled: boolean;
  timeout: number;
}

export function DefaultAgentConfig(): AgentConfig {
  return {
    agentId: '',
    name: 'Node.js Agent',
    centerUrl: 'localhost:9090',
    connectionType: 'http',
    heartbeatInterval: 30,
    reconnectInterval: 5,
    maxReconnectAttempts: 10,
    skills: [],
    metadata: {},
    tlsEnabled: false,
    timeout: 30,
  };
}