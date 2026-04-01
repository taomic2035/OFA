/**
 * OFA Node.js SDK - P2P 模块
 * Agent 间直接通信
 */

import * as dgram from 'dgram';
import * as net from 'net';
import { v4 as uuidv4 } from 'uuid';

/**
 * P2P 消息类型
 */
export enum P2PMessageType {
  DATA = 'data',
  BROADCAST = 'broadcast',
  REQUEST = 'request',
  RESPONSE = 'response',
  DISCOVERY = 'discovery',
  HEARTBEAT = 'heartbeat',
}

/**
 * P2P 消息
 */
export interface P2PMessage {
  type: P2PMessageType;
  fromId: string;
  toId?: string;
  data?: unknown;
  timestamp: number;
  msgId: string;
}

/**
 * 设备信息
 */
export interface PeerInfo {
  id: string;
  name: string;
  address: string;
  port: number;
  online: boolean;
  lastSeen: number;
  latencyMs: number;
}

/**
 * P2P 客户端
 */
export class P2PClient {
  private agentId: string;
  private port: number;
  private peers: Map<string, PeerInfo> = new Map();
  private running = false;

  private server?: net.Server;
  private discoverySocket?: dgram.Socket;

  private messageHandlers: Array<(msg: P2PMessage) => void> = [];
  private peerHandlers: Array<(peerId: string, peerName: string, added: boolean) => void> = [];

  constructor(agentId: string, port: number = 0) {
    this.agentId = agentId;
    this.port = port || 9000 + Math.floor(Math.random() * 1000);
  }

  /**
   * 启动 P2P 服务
   */
  async start(): Promise<void> {
    if (this.running) return;
    this.running = true;

    // 启动 TCP 服务器
    await this.startServer();

    // 启动 UDP 发现
    this.startDiscovery();

    console.log(`P2P client started on port ${this.port}`);
  }

  /**
   * 停止 P2P 服务
   */
  stop(): void {
    this.running = false;

    if (this.server) {
      this.server.close();
    }

    if (this.discoverySocket) {
      this.discoverySocket.close();
    }

    console.log('P2P client stopped');
  }

  private async startServer(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.server = net.createServer((socket) => {
        let data = Buffer.alloc(0);

        socket.on('data', (chunk) => {
          data = Buffer.concat([data, chunk]);
        });

        socket.on('end', () => {
          try {
            const msg: P2PMessage = JSON.parse(data.toString());
            this.handleMessage(msg);
          } catch (error) {
            console.error('Invalid message:', error);
          }
        });

        socket.on('error', (error) => {
          console.error('Socket error:', error);
        });
      });

      this.server.listen(this.port, () => {
        resolve();
      });

      this.server.on('error', reject);
    });
  }

  private startDiscovery(): void {
    this.discoverySocket = dgram.createSocket('udp4');
    this.discoverySocket.bind(9999, () => {
      this.discoverySocket?.setBroadcast(true);
    });

    this.discoverySocket.on('message', (msg, rinfo) => {
      try {
        const data = JSON.parse(msg.toString());
        if (data.type === 'discovery' && data.agentId !== this.agentId) {
          const peer: PeerInfo = {
            id: data.agentId,
            name: data.agentId,
            address: rinfo.address,
            port: data.port,
            online: true,
            lastSeen: Date.now(),
            latencyMs: 0,
          };
          this.addPeer(peer);
        }
      } catch (error) {
        // 忽略无效消息
      }
    });

    // 定时广播
    setInterval(() => {
      if (!this.running) return;

      const discoveryMsg = JSON.stringify({
        type: 'discovery',
        agentId: this.agentId,
        port: this.port,
        timestamp: Date.now(),
      });

      this.discoverySocket?.send(
        discoveryMsg,
        9999,
        '255.255.255.255',
        (error) => {
          if (error) console.error('Broadcast error:', error);
        }
      );
    }, 5000);
  }

  private handleMessage(msg: P2PMessage): void {
    // 更新设备状态
    const peer = this.peers.get(msg.fromId);
    if (peer) {
      peer.lastSeen = Date.now();
      peer.online = true;
    }

    // 通知处理器
    for (const handler of this.messageHandlers) {
      try {
        handler(msg);
      } catch (error) {
        console.error('Message handler error:', error);
      }
    }
  }

  /**
   * 发送消息
   */
  async send(peerId: string, data: unknown): Promise<boolean> {
    const peer = this.peers.get(peerId);
    if (!peer || !peer.online) {
      console.warn(`Peer not found or offline: ${peerId}`);
      return false;
    }

    const msg: P2PMessage = {
      type: P2PMessageType.DATA,
      fromId: this.agentId,
      toId: peerId,
      data,
      timestamp: Date.now(),
      msgId: uuidv4().split('-')[0],
    };

    return this.sendToPeer(peer, msg);
  }

  private sendToPeer(peer: PeerInfo, msg: P2PMessage): Promise<boolean> {
    return new Promise((resolve) => {
      const socket = net.createConnection({ host: peer.address, port: peer.port }, () => {
        socket.end(JSON.stringify(msg));
        resolve(true);
      });

      socket.on('error', (error) => {
        console.error(`Send to ${peer.id} failed:`, error);
        peer.online = false;
        resolve(false);
      });

      socket.setTimeout(5000, () => {
        socket.destroy();
        resolve(false);
      });
    });
  }

  /**
   * 广播消息
   */
  async broadcast(data: unknown): Promise<Record<string, boolean>> {
    const results: Record<string, boolean> = {};

    for (const [peerId, peer] of this.peers) {
      if (peer.online) {
        results[peerId] = await this.send(peerId, data);
      }
    }

    return results;
  }

  /**
   * 添加设备
   */
  addPeer(peer: PeerInfo): void {
    this.peers.set(peer.id, peer);

    for (const handler of this.peerHandlers) {
      handler(peer.id, peer.name, true);
    }

    console.log(`Peer added: ${peer.id}`);
  }

  /**
   * 移除设备
   */
  removePeer(peerId: string): void {
    const peer = this.peers.get(peerId);
    if (peer) {
      this.peers.delete(peerId);

      for (const handler of this.peerHandlers) {
        handler(peerId, peer.name, false);
      }

      console.log(`Peer removed: ${peerId}`);
    }
  }

  /**
   * 获取设备列表
   */
  getPeers(): PeerInfo[] {
    return Array.from(this.peers.values());
  }

  /**
   * 获取在线设备
   */
  getOnlinePeers(): PeerInfo[] {
    return Array.from(this.peers.values()).filter(p => p.online);
  }

  /**
   * 获取端口
   */
  getPort(): number {
    return this.port;
  }

  /**
   * 添加消息处理器
   */
  onMessage(handler: (msg: P2PMessage) => void): void {
    this.messageHandlers.push(handler);
  }

  /**
   * 添加设备变化处理器
   */
  onPeerChange(handler: (peerId: string, peerName: string, added: boolean) => void): void {
    this.peerHandlers.push(handler);
  }

  /**
   * 检查设备状态
   */
  checkPeersStatus(): void {
    const now = Date.now();
    for (const [id, peer] of this.peers) {
      if (now - peer.lastSeen > 30000) {
        peer.online = false;
        for (const handler of this.peerHandlers) {
          handler(id, peer.name, false);
        }
      }
    }
  }

  /**
   * 获取统计信息
   */
  getStats(): { agentId: string; port: number; peersTotal: number; peersOnline: number } {
    const online = Array.from(this.peers.values()).filter(p => p.online).length;
    return {
      agentId: this.agentId,
      port: this.port,
      peersTotal: this.peers.size,
      peersOnline: online,
    };
  }
}