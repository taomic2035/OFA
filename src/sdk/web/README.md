# OFA Web SDK

Version: 8.3.0

OFA Web SDK provides browser-based agent implementation for web applications.

## Features

- **WebSocket Connection**: Connect to OFA Center server from browser
- **Identity Management**: Personal identity sync and behavior observation
- **Scene Detection**: Detect user context from browser signals
- **Audio Playback**: TTS and voice streaming playback using Web Audio API
- **Chat Client**: LLM dialogue integration with streaming support
- **TypeScript**: Full TypeScript support with type definitions

## Requirements

- Modern browser with WebSocket support
- ES2020+ JavaScript support
- Web Audio API (for audio playback)

## Installation

### NPM

```bash
npm install ofa-web-sdk
```

### CDN

```html
<script src="https://cdn.jsdelivr.net/npm/ofa-web-sdk/dist/index.js"></script>
```

### Direct Import

```typescript
import { OFAWebAgent } from './path/to/ofa-web-sdk/src';
```

## Usage

### Basic Setup

```typescript
import { OFAWebAgent, AgentConfig } from 'ofa-web-sdk';

const config: AgentConfig = {
  centerAddress: 'ws://your-center:8080/ws',
  mode: 'sync'
};

const agent = new OFAWebAgent(config);

// Initialize
await agent.initialize();

// Connect to Center
await agent.connectCenter();
```

### Identity Management

```typescript
// Create identity
const identity = await agent.identityManager.createIdentity('User');

// Observe behavior
agent.identityManager.observeBehavior('decision', {
  impulse_purchase: true
});

// Trigger personality inference
await agent.identityManager.inferPersonality();

// Get decision context
const context = agent.identityManager.getDecisionContext();
```

### Scene Detection

```typescript
// Initialize scene detector
agent.sceneDetector.initialize();

// Manual detection
await agent.sceneDetector.detect();

// Listen for scene changes
agent.sceneDetector.on('sceneStart', (scene) => {
  console.log('Scene started:', scene.type);
});
```

### Audio Playback

```typescript
// Initialize audio player
agent.audioPlayer.initialize();

// Play audio
const audioBuffer = await fetch('/audio.wav').then(r => r.arrayBuffer());
await agent.audioPlayer.play(audioBuffer);

// Streaming playback
agent.audioPlayer.playStream();
agent.audioPlayer.queueAudioChunk(chunk);

// Control
agent.audioPlayer.pause();
agent.audioPlayer.resume();
agent.audioPlayer.stop();
```

### Chat Integration

```typescript
// Send chat message
const response = await agent.chatClient.chat({
  message: 'Hello',
  sessionId: 'my_session'
});

// Streaming chat
await agent.chatClient.chatStream(
  { message: 'Tell me a story' },
  (chunk) => console.log(chunk)
);

// TTS request
const audioData = await agent.chatClient.requestTTS('Hello world');
await agent.audioPlayer.play(audioData);
```

## Architecture

### Core Components

| Component | Description |
|-----------|-------------|
| OFAWebAgent | Main entry point for browser |
| CenterConnection | WebSocket connection to Center |
| IdentityManager | Identity sync and behavior observation |
| SceneDetector | Scene detection from browser |
| AudioPlayer | Audio playback using Web Audio API |
| ChatClient | LLM dialogue integration |

### Browser Capabilities

The SDK detects browser capabilities automatically:

| Capability | Detection |
|------------|-----------|
| voice | Always available |
| display | Always available |
| camera | navigator.mediaDevices |
| audio_input | navigator.mediaDevices |
| location | navigator.geolocation |
| keyboard | Always available |
| mouse | Always available |

## Events

All components use EventEmitter3 for event handling:

```typescript
// Connection events
agent.connection.on('connected', () => { ... });
agent.connection.on('message', (msg) => { ... });

// Identity events
agent.identityManager.on('identityChange', (id) => { ... });
agent.identityManager.on('personalityUpdated', (p) => { ... });

// Scene events
agent.sceneDetector.on('sceneStart', (scene) => { ... });
agent.sceneDetector.on('sceneEnd', (scene) => { ... });

// Audio events
agent.audioPlayer.on('stateChange', (state) => { ... });

// Chat events
agent.chatClient.on('response', (resp) => { ... });
agent.chatClient.on('streamChunk', (chunk) => { ... });
```

## Storage

The SDK uses localStorage for identity persistence:

- Key: `ofa_identity`
- Content: JSON-encoded PersonalIdentity

## License

MIT License
