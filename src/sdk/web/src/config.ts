/**
 * OFA Web SDK - Configuration (v8.3.0)
 */

export interface AgentConfig {
  agentId?: string;
  centerAddress?: string;
  mode?: 'standalone' | 'sync';
  heartbeatInterval?: number;
  syncInterval?: number;
  enableCache?: boolean;
  autoConnect?: boolean;
  debug?: boolean;
}

export const defaultConfig: AgentConfig = {
  mode: 'sync',
  heartbeatInterval: 30000,
  syncInterval: 300000,
  enableCache: true,
  autoConnect: false,
  debug: false
};
