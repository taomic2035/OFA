/**
 * OFA Web SDK - WebSocket Connection (v8.3.0)
 */

import EventEmitter from 'eventemitter3';
import {
  ConnectionState,
  MessageType,
  WebSocketMessage,
  AgentProfile,
  AgentStatus,
  OFAError,
  OFAErrorCode
} from './types';

export interface ConnectionEvents {
  connected: [];
  disconnected: [];
  message: [WebSocketMessage];
  error: [Error];
  stateChange: [ConnectionState];
}

export class CenterConnection extends EventEmitter<ConnectionEvents> {
  private ws: WebSocket | null = null;
  private centerAddress: string;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts: number = 3;
  private heartbeatTimer: number | null = null;
  private _state: ConnectionState = ConnectionState.Disconnected;
  private _sessionId: string | null = null;

  constructor(centerAddress?: string) {
    super();
    this.centerAddress = centerAddress || '';
  }

  get state(): ConnectionState {
    return this._state;
  }

  get sessionId(): string | null {
    return this._sessionId;
  }

  async connect(address: string): Promise<void> {
    this.centerAddress = address;
    this._state = ConnectionState.Connecting;
    this.emit('stateChange', this._state);

    try {
      this.ws = new WebSocket(address);
      
      this.ws.onopen = () => {
        this._state = ConnectionState.Connected;
        this.reconnectAttempts = 0;
        this.emit('stateChange', this._state);
        this.emit('connected');
        this.startReceiving();
      };

      this.ws.onclose = () => {
        this.handleDisconnect();
      };

      this.ws.onerror = (error) => {
        this.handleError(error);
      };

    } catch (error) {
      this._state = ConnectionState.Error;
      this.emit('stateChange', this._state);
      throw new OFAError(OFAErrorCode.ConnectionError, `Connection failed: ${error}`);
    }
  }

  async register(profile: AgentProfile): Promise<void> {
    const message: WebSocketMessage = {
      type: MessageType.Register,
      payload: {
        agent_id: profile.agentId,
        device_type: profile.deviceType,
        device_name: profile.deviceName,
        browser_info: profile.browserInfo,
        capabilities: profile.capabilities,
        identity_id: profile.identityId || ''
      },
      timestamp: new Date()
    };
    await this.send(message);
  }

  async sendHeartbeat(profile: AgentProfile): Promise<void> {
    const message: WebSocketMessage = {
      type: MessageType.Heartbeat,
      payload: {
        agent_id: profile.agentId,
        status: profile.status,
        timestamp: new Date().toISOString()
      },
      timestamp: new Date()
    };
    await this.send(message);
  }

  async send(message: WebSocketMessage): Promise<void> {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new OFAError(OFAErrorCode.ConnectionError, 'WebSocket not connected');
    }
    this.ws.send(JSON.stringify(message));
  }

  disconnect(): void {
    this.stopHeartbeat();
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this._state = ConnectionState.Disconnected;
    this._sessionId = null;
    this.emit('stateChange', this._state);
    this.emit('disconnected');
  }

  private startReceiving(): void {
    if (!this.ws) return;

    this.ws.onmessage = (event) => {
      try {
        const message: WebSocketMessage = JSON.parse(event.data);
        this.handleMessage(message);
      } catch (error) {
        console.error('Failed to parse message:', error);
      }
    };
  }

  private handleMessage(message: WebSocketMessage): void {
    this.emit('message', message);

    switch (message.type) {
      case MessageType.RegisterAck:
        if (message.payload.session_id) {
          this._sessionId = message.payload.session_id;
        }
        break;

      case MessageType.Error:
        const errorMsg = message.payload.message || 'Unknown error';
        this.emit('error', new OFAError(OFAErrorCode.ConnectionError, errorMsg));
        break;
    }
  }

  private handleDisconnect(): void {
    this._state = ConnectionState.Disconnected;
    this.emit('stateChange', this._state);
    this.emit('disconnected');

    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.attemptReconnect();
    }
  }

  private handleError(error: any): void {
    this._state = ConnectionState.Error;
    this.emit('stateChange', this._state);
    this.emit('error', error);
  }

  private attemptReconnect(): void {
    this.reconnectAttempts++;
    this._state = ConnectionState.Reconnecting;
    this.emit('stateChange', this._state);

    setTimeout(() => {
      this.connect(this.centerAddress).catch(() => {});
    }, 5000);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  isConnected(): boolean {
    return this._state === ConnectionState.Connected;
  }
}
