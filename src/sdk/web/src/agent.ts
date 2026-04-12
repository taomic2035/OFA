/**
 * OFA Web SDK - Web Agent Implementation (v8.3.0)
 */

import EventEmitter from 'eventemitter3';
import { CenterConnection } from './connection';
import { IdentityManager } from './identity';
import { SceneDetector } from './scene';
import { AudioPlayer } from './audio';
import { ChatClient } from './chat';
import {
  AgentStatus,
  AgentMode,
  DeviceType,
  AgentProfile,
  ConnectionState,
  OFAError,
  OFAErrorCode
} from './types';
import { AgentConfig, defaultConfig } from './config';

export interface AgentEvents {
  statusChange: [AgentStatus];
  connected: [];
  disconnected: [];
  identityChange: [string | null];
}

export class OFAWebAgent extends EventEmitter<AgentEvents> {
  private config: AgentConfig;
  private _status: AgentStatus = AgentStatus.Offline;
  private _connectedToCenter: boolean = false;
  private _currentIdentityId: string | null = null;

  public readonly profile: AgentProfile;
  public readonly connection: CenterConnection;
  public readonly identityManager: IdentityManager;
  public readonly sceneDetector: SceneDetector;
  public readonly audioPlayer: AudioPlayer;
  public readonly chatClient: ChatClient;

  private heartbeatTimer: number | null = null;

  constructor(config: AgentConfig = {}) {
    super();
    this.config = { ...defaultConfig, ...config };

    const deviceType = this.detectDeviceType();
    const browserInfo = this.getBrowserInfo();

    this.profile = {
      agentId: this.config.agentId || this.generateAgentId(),
      deviceType,
      deviceName: this.getDeviceName(),
      browserInfo,
      capabilities: this.getCapabilities(deviceType),
      status: AgentStatus.Offline,
      lastHeartbeat: new Date()
    };

    this.connection = new CenterConnection(this.config.centerAddress);
    this.identityManager = new IdentityManager();
    this.sceneDetector = new SceneDetector();
    this.audioPlayer = new AudioPlayer();
    this.chatClient = new ChatClient();

    this.setupBindings();
  }

  get status(): AgentStatus {
    return this._status;
  }

  get connectedToCenter(): boolean {
    return this._connectedToCenter;
  }

  get currentIdentityId(): string | null {
    return this._currentIdentityId;
  }

  async initialize(): Promise<void> {
    await this.identityManager.initialize();
    this.audioPlayer.initialize();

    if (this.config.mode === 'sync' && this.config.centerAddress) {
      await this.connectCenter();
    }

    this._status = AgentStatus.Online;
    this.emit('statusChange', this._status);
  }

  async connectCenter(): Promise<void> {
    if (!this.config.centerAddress) {
      throw new OFAError(OFAErrorCode.ConfigurationError, 'Center address not configured');
    }

    await this.connection.connect(this.config.centerAddress);
    await this.connection.register(this.profile);

    this._connectedToCenter = true;
    this._status = AgentStatus.Online;
    this.emit('statusChange', this._status);
    this.emit('connected');

    this.startHeartbeat();
  }

  async disconnect(): Promise<void> {
    this.connection.disconnect();
    this.stopHeartbeat();
    this._connectedToCenter = false;
    this._status = AgentStatus.Offline;
    this.emit('statusChange', this._status);
    this.emit('disconnected');
  }

  async syncWithCenter(): Promise<void> {
    if (!this._connectedToCenter) {
      throw new OFAError(OFAErrorCode.ConnectionError, 'Not connected to Center');
    }

    const identity = this.identityManager.currentIdentity;
    if (identity) {
      await this.connection.send({
        type: 'SyncRequest' as any,
        payload: { identity },
        timestamp: new Date()
      });
    }

    const behaviors = await this.identityManager.getPendingBehaviors();
    if (behaviors.length > 0) {
      await this.connection.send({
        type: 'BehaviorReport' as any,
        payload: { behaviors },
        timestamp: new Date()
      });
    }
  }

  setMode(mode: AgentMode): void {
    this.config.mode = mode === AgentMode.Sync ? 'sync' : 'standalone';

    if (this.config.mode === 'sync' && !this._connectedToCenter && this.config.centerAddress) {
      this.connectCenter().catch(() => {});
    }
  }

  private detectDeviceType(): DeviceType {
    const userAgent = navigator.userAgent.toLowerCase();
    if (/mobile|android|iphone|ipad|ipod/.test(userAgent)) {
      return DeviceType.Mobile;
    } else if (/tablet|ipad/.test(userAgent)) {
      return DeviceType.Mobile;
    }
    return DeviceType.Browser;
  }

  private getDeviceName(): string {
    return navigator.userAgent.split(' ').slice(-1).join(' ') || 'Web Browser';
  }

  private getBrowserInfo(): string {
    return navigator.userAgent;
  }

  private getCapabilities(deviceType: DeviceType): string[] {
    const capabilities = ['voice', 'display'];
    
    if (navigator.mediaDevices) {
      capabilities.push('camera', 'audio_input');
    }
    
    if ('geolocation' in navigator) {
      capabilities.push('location');
    }
    
    capabilities.push('keyboard', 'mouse');
    
    return capabilities;
  }

  private generateAgentId(): string {
    return 'agent_web_' + Math.random().toString(36).substring(2, 15);
  }

  private setupBindings(): void {
    this.connection.on('stateChange', (state) => {
      if (state === ConnectionState.Connected) {
        this._connectedToCenter = true;
      } else {
        this._connectedToCenter = false;
      }
    });

    this.connection.on('disconnected', () => {
      this._connectedToCenter = false;
      this.emit('disconnected');
    });

    this.identityManager.on('identityChange', (id) => {
      this._currentIdentityId = id;
      this.emit('identityChange', id);
    });
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.connection.isConnected()) {
        this.connection.sendHeartbeat(this.profile).catch(() => {});
      }
    }, this.config.heartbeatInterval!);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  getStatus(): AgentStatus {
    return this._status;
  }
}
