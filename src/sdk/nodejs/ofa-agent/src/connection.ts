/**
 * Connection Module
 * Sprint 29: Node.js Agent SDK
 */

import WebSocket from 'ws';
import axios from 'axios';
import { AgentConfig } from './config';
import { Message } from './message';

/**
 * 连接接口
 */
export interface Connection {
  connect(): Promise<void>;
  disconnect(): Promise<void>;
  send(msg: Message): Promise<void>;
  receive(): Promise<Message | null>;
  isConnected(): boolean;
}

/**
 * 创建连接
 */
export function createConnection(config: AgentConfig): Connection {
  switch (config.connectionType) {
    case 'http':
      return new HttpConnection(config);
    case 'websocket':
      return new WebSocketConnection(config);
    case 'grpc':
      return new GrpcConnection(config);
    default:
      return new HttpConnection(config);
  }
}

/**
 * 连接类型
 */
export enum ConnectionType {
  HTTP = 'http',
  WEBSOCKET = 'websocket',
  GRPC = 'grpc',
}

/**
 * HTTP连接
 */
export class HttpConnection implements Connection {
  private config: AgentConfig;
  private connected: boolean = false;
  private baseUrl: string = '';
  private messageQueue: Message[] = [];

  constructor(config: AgentConfig) {
    this.config = config;
  }

  async connect(): Promise<void> {
    this.baseUrl = `http://${this.config.centerUrl}`;
    this.connected = true;

    // 注册Agent
    await axios.post(`${this.baseUrl}/api/v1/agents/register`, {
      id: this.config.agentId,
      name: this.config.name,
      type: 'nodejs',
    });

    // 启动消息轮询
    this.startPolling();
  }

  async disconnect(): Promise<void> {
    this.connected = false;
  }

  async send(msg: Message): Promise<void> {
    if (!this.connected) throw new Error('Not connected');

    const type = msg.type.toLowerCase();
    await axios.post(`${this.baseUrl}/api/v1/${type}`, msg);
  }

  async receive(): Promise<Message | null> {
    return this.messageQueue.shift() || null;
  }

  isConnected(): boolean {
    return this.connected;
  }

  private startPolling(): void {
    const poll = async () => {
      if (!this.connected) return;

      try {
        const response = await axios.get(
          `${this.baseUrl}/api/v1/agents/${this.config.agentId}/messages`
        );

        for (const msg of response.data) {
          this.messageQueue.push(msg);
        }
      } catch (error) {
        console.error('Polling error:', error);
      }

      setTimeout(poll, 1000);
    };

    poll();
  }
}

/**
 * WebSocket连接
 */
export class WebSocketConnection implements Connection {
  private config: AgentConfig;
  private ws: WebSocket | null = null;
  private connected: boolean = false;
  private messageQueue: Message[] = [];

  constructor(config: AgentConfig) {
    this.config = config;
  }

  async connect(): Promise<void> {
    const url = `ws://${this.config.centerUrl}/ws`;

    return new Promise((resolve, reject) => {
      this.ws = new WebSocket(url);

      this.ws.on('open', () => {
        this.connected = true;
        resolve();
      });

      this.ws.on('message', (data: Buffer) => {
        try {
          const msg = JSON.parse(data.toString()) as Message;
          this.messageQueue.push(msg);
        } catch (error) {
          console.error('Parse error:', error);
        }
      });

      this.ws.on('close', () => {
        this.connected = false;
      });

      this.ws.on('error', (error: Error) => {
        reject(error);
      });
    });
  }

  async disconnect(): Promise<void> {
    this.connected = false;
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  async send(msg: Message): Promise<void> {
    if (!this.ws || !this.connected) throw new Error('Not connected');
    this.ws.send(JSON.stringify(msg));
  }

  async receive(): Promise<Message | null> {
    return this.messageQueue.shift() || null;
  }

  isConnected(): boolean {
    return this.connected;
  }
}

/**
 * gRPC连接（模拟）
 */
export class GrpcConnection implements Connection {
  private config: AgentConfig;
  private connected: boolean = false;

  constructor(config: AgentConfig) {
    this.config = config;
  }

  async connect(): Promise<void> {
    // 实际实现需要@grpc/grpc-js
    this.connected = true;
  }

  async disconnect(): Promise<void> {
    this.connected = false;
  }

  async send(_msg: Message): Promise<void> {
    // gRPC发送
  }

  async receive(): Promise<Message | null> {
    return null;
  }

  isConnected(): boolean {
    return this.connected;
  }
}