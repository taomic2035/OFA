# OFA API 文档

## 概述

OFA 提供两种 API 接口：
- **REST API**: HTTP 接口，端口 8080
- **gRPC API**: 高性能 RPC 接口，端口 9090
- **状态推送**: Center 向设备端推送状态更新

---

## 版本概览

| 版本系列 | 特性 | API 命名空间 |
|---------|------|-------------|
| v2.x | 去中心化架构 | `/api/v1/identities`, `/api/v1/sync`, `/api/v1/devices` |
| v3.x | 多设备协同 | `/api/v1/agents`, `/api/v1/tasks`, `/api/v1/messages` |
| v4.x | 灵魂特征 | `/api/v1/emotions`, `/api/v1/philosophy`, `/api/v1/social`, `/api/v1/culture`, `/api/v1/lifestage`, `/api/v1/relationship` |
| v5.x | 外在呈现 | `/api/v1/avatar`, `/api/v1/expression`, `/api/v1/speech`, `/api/v1/tts` |
| v6.x | REST API 完善 | 统一入口 `pkg/rest/server.go` |

---

## REST API

### 基础信息

- 基础 URL: `http://localhost:8080`
- 内容类型: `application/json`
- 认证: 无 (开发版本)

---

## 基础 API

### 健康检查

```
GET /health
```

**响应:**

```json
{
  "status": "healthy",
  "version": "v6.3.0"
}
```

---

## v2.x 分布式架构 API

### 身份管理 (Identity)

#### 创建身份

```
POST /api/v1/identities
```

**请求体:**

```json
{
  "id": "user-123",
  "name": "张三",
  "nickname": "小张",
  "avatar": "https://example.com/avatar.jpg",
  "personality": {
    "openness": 0.7,
    "conscientiousness": 0.8,
    "extraversion": 0.6,
    "agreeableness": 0.75,
    "neuroticism": 0.3
  }
}
```

#### 获取身份列表

```
GET /api/v1/identities?page=1&page_size=20
```

#### 获取单个身份

```
GET /api/v1/identities/{id}
```

#### 更新身份

```
PUT /api/v1/identities/{id}
```

#### 删除身份

```
DELETE /api/v1/identities/{id}
```

---

### 设备管理 (Device)

#### 注册设备

```
POST /api/v1/devices
```

**请求体:**

```json
{
  "agent_id": "device-001",
  "identity_id": "user-123",
  "device_type": "phone",
  "device_name": "iPhone 15",
  "capabilities": ["voice", "display", "camera"]
}
```

#### 获取设备列表

```
GET /api/v1/devices?identity_id=user-123
```

#### 获取单个设备

```
GET /api/v1/devices/{id}
```

#### 设备心跳

```
POST /api/v1/devices/{id}/heartbeat
```

**请求体:**

```json
{
  "status": "online",
  "battery": 85,
  "network": "wifi"
}
```

---

### 行为上报 (Behavior)

#### 上报行为

```
POST /api/v1/behaviors
```

**请求体:**

```json
{
  "agent_id": "device-001",
  "identity_id": "user-123",
  "type": "decision",
  "observation": {
    "action": "purchase",
    "item": "coffee",
    "impulse": true
  }
}
```

#### 获取行为列表

```
GET /api/v1/behaviors/{identity_id}
```

---

### 数据同步 (Sync)

#### 同步数据

```
POST /api/v1/sync
```

**请求体:**

```json
{
  "agent_id": "device-001",
  "identity_id": "user-123",
  "sync_type": "full",
  "changes": []
}
```

#### 获取同步状态

```
GET /api/v1/sync/{identity_id}/state
```

---

## v3.x 多设备协同 API

### Agent 管理

#### 获取 Agent 列表

```
GET /api/v1/agents
```

#### 获取单个 Agent

```
GET /api/v1/agents/{id}
```

#### 删除 Agent

```
DELETE /api/v1/agents/{id}
```

---

### 任务管理 (Task)

#### 提交任务

```
POST /api/v1/tasks
```

**请求体:**

```json
{
  "skill_id": "text.process",
  "input": "base64_encoded_input",
  "target_agent": "",
  "priority": 0,
  "timeout_ms": 30000
}
```

#### 获取任务列表

```
GET /api/v1/tasks
```

#### 获取任务状态

```
GET /api/v1/tasks/{id}
```

#### 取消任务

```
POST /api/v1/tasks/{id}/cancel
```

---

### 消息通信 (Message)

#### 发送消息

```
POST /api/v1/messages
```

#### 广播消息

```
POST /api/v1/messages/broadcast
```

#### 组播消息

```
POST /api/v1/messages/multicast
```

---

### 技能管理

#### 获取技能列表

```
GET /api/v1/skills
```

---

### 系统信息

#### 获取系统信息

```
GET /api/v1/system/info
```

#### 获取系统指标

```
GET /api/v1/system/metrics
```

---

## v4.x 灵魂特征 API

### 情绪系统 (Emotion) - v4.0.0

#### 触发情绪

```
POST /api/v1/emotions/trigger
```

**请求体:**

```json
{
  "identity_id": "user-123",
  "trigger_type": "event",
  "trigger_desc": "Received a compliment from friend",
  "emotion_type": "joy",
  "intensity": 0.8
}
```

#### 获取情绪状态

```
GET /api/v1/emotions/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "current_emotion": {
    "joy": 0.6,
    "anger": 0.1,
    "sadness": 0.1,
    "fear": 0.1,
    "love": 0.7,
    "disgust": 0.05,
    "desire": 0.3
  },
  "dominant_emotion": "joy",
  "emotion_intensity": 0.6,
  "emotion_stability": 0.7
}
```

#### 获取情绪上下文

```
GET /api/v1/emotions/{identity_id}/context
```

#### 获取情绪画像

```
GET /api/v1/emotions/{identity_id}/profile
```

**响应:**

```json
{
  "identity_id": "user-123",
  "base_joy_level": 0.5,
  "base_anger_level": 0.1,
  "base_sadness_level": 0.1,
  "base_fear_level": 0.1,
  "base_love_level": 0.4,
  "base_disgust_level": 0.05,
  "base_desire_level": 0.3,
  "emotional_stability": 0.7
}
```

#### 更新情绪画像

```
PUT /api/v1/emotions/{identity_id}/profile
```

---

### 三观系统 (Philosophy) - v4.1.0

#### 设置世界观

```
POST /api/v1/philosophy/worldview
```

**请求体:**

```json
{
  "identity_id": "user-123",
  "optimism": 0.7,
  "change_belief": 0.6,
  "trust_in_people": 0.65,
  "fate_control": 0.5,
  "world_essence": "material",
  "society_view": "fair",
  "future_view": "optimistic",
  "relationship_view": "cooperative"
}
```

#### 获取世界观

```
GET /api/v1/philosophy/{identity_id}/worldview
```

**响应:**

```json
{
  "identity_id": "user-123",
  "optimism": 0.7,
  "change_belief": 0.6,
  "trust_in_people": 0.65,
  "fate_control": 0.5,
  "world_essence": "material",
  "society_view": "fair",
  "future_view": "optimistic",
  "relationship_view": "cooperative"
}
```

#### 获取三观上下文

```
GET /api/v1/philosophy/{identity_id}/context
```

---

### 社会身份 (Social) - v4.2.0

#### 获取社会身份

```
GET /api/v1/social/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "education": {
    "level": "bachelor",
    "field": "computer_science",
    "school_tier": "top_50",
    "continuous_learning": true
  },
  "career": {
    "industry": "technology",
    "occupation": "software_engineer",
    "stage": "early_career",
    "satisfaction": 0.7
  },
  "social_class": {
    "income_level": "middle",
    "wealth_level": "moderate",
    "cultural_capital": 0.6,
    "social_capital": 0.5,
    "economic_capital": 0.55
  }
}
```

#### 更新社会身份

```
PUT /api/v1/social/{identity_id}
```

#### 获取教育背景

```
GET /api/v1/social/{identity_id}/education
```

#### 更新教育背景

```
PUT /api/v1/social/{identity_id}/education
```

#### 获取职业画像

```
GET /api/v1/social/{identity_id}/career
```

#### 更新职业画像

```
PUT /api/v1/social/{identity_id}/career
```

#### 获取社会上下文

```
GET /api/v1/social/{identity_id}/context
```

---

### 地域文化 (Culture) - v4.3.0

#### 获取地域文化

```
GET /api/v1/culture/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "region": {
    "province": "guangdong",
    "city": "shenzhen",
    "city_tier": 1,
    "region_zone": "south_china"
  },
  "cultural_dimensions": {
    "collectivism": 0.65,
    "tradition_orientation": 0.55,
    "power_distance": 0.5,
    "uncertainty_avoidance": 0.45,
    "long_term_orientation": 0.7
  },
  "communication_style": "direct",
  "social_style": "open"
}
```

#### 更新地域文化

```
PUT /api/v1/culture/{identity_id}
```

#### 设置位置

```
POST /api/v1/culture/{identity_id}/location
```

**请求体:**

```json
{
  "province": "guangdong",
  "city": "shenzhen",
  "city_tier": 1
}
```

#### 获取文化上下文

```
GET /api/v1/culture/{identity_id}/context
```

---

### 人生阶段 (LifeStage) - v4.4.0

#### 获取人生阶段

```
GET /api/v1/lifestage/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "current_stage": "early_adulthood",
  "stage_progress": 0.3,
  "life_events": [
    {
      "event_type": "career",
      "event_name": "first_job",
      "impact": 0.8,
      "occurred_at": "2020-06-01"
    }
  ],
  "development_metrics": {
    "physical": 0.9,
    "cognitive": 0.8,
    "emotional": 0.75,
    "social": 0.7,
    "professional": 0.6
  }
}
```

#### 更新人生阶段

```
PUT /api/v1/lifestage/{identity_id}
```

#### 设置当前阶段

```
POST /api/v1/lifestage/{identity_id}/stage
```

**请求体:**

```json
{
  "stage_name": "early_adulthood",
  "age": 25
}
```

#### 添加人生事件

```
POST /api/v1/lifestage/{identity_id}/event
```

**请求体:**

```json
{
  "event_type": "career",
  "event_name": "job_change",
  "impact": 0.6,
  "description": "Changed to a new company"
}
```

#### 获取人生阶段上下文

```
GET /api/v1/lifestage/{identity_id}/context
```

---

### 人际关系 (Relationship) - v4.6.0

#### 获取人际关系系统

```
GET /api/v1/relationship/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "social_network": {
    "network_size": 150,
    "strong_ties": 5,
    "weak_ties": 30,
    "network_density": 0.3,
    "diversity_score": 0.6
  },
  "attachment_style": "secure",
  "relationship_profile": {
    "relationship_tendency": "balanced",
    "social_style": {
      "directness": 0.6,
      "expressiveness": 0.7
    },
    "conflict_style": "cooperative",
    "relationship_capacity": 0.75
  }
}
```

#### 更新人际关系系统

```
PUT /api/v1/relationship/{identity_id}
```

#### 添加关系

```
POST /api/v1/relationship/{identity_id}/add
```

**请求体:**

```json
{
  "person_id": "friend-001",
  "person_name": "李四",
  "relationship_type": "friend",
  "intimacy": 0.7,
  "trust": 0.8,
  "importance": 0.75
}
```

#### 获取人际关系上下文

```
GET /api/v1/relationship/{identity_id}/context
```

---

## v5.x 外在呈现 API

### 外在形象 (Avatar) - v5.0.0

#### 获取 Avatar

```
GET /api/v1/avatar/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "facial_features": {
    "face_shape": "oval",
    "eye_shape": "almond",
    "eye_color": "brown",
    "skin_tone": "light",
    "hair_style": "short",
    "hair_color": "black"
  },
  "body_features": {
    "body_type": "average",
    "height_category": "tall",
    "build": "fit",
    "metabolism": "normal",
    "posture_style": "confident"
  },
  "style_preferences": {
    "clothing_style": "casual",
    "color_preferences": ["blue", "black"],
    "accessories_style": "minimal",
    "grooming_style": "neat",
    "fashion_trendiness": 0.4
  },
  "age_appearance": {
    "apparent_age": 28,
    "aging_stage": "prime",
    "facial_maturity": 0.6
  }
}
```

#### 更新 Avatar

```
PUT /api/v1/avatar/{identity_id}
```

#### 更新面部特征

```
PUT /api/v1/avatar/{identity_id}/facial
```

**请求体:**

```json
{
  "face_shape": "oval",
  "eye_shape": "almond",
  "hair_style": "medium"
}
```

#### 更新身体特征

```
PUT /api/v1/avatar/{identity_id}/body
```

**请求体:**

```json
{
  "body_type": "fit",
  "posture_style": "confident"
}
```

#### 更新风格偏好

```
PUT /api/v1/avatar/{identity_id}/style
```

**请求体:**

```json
{
  "clothing_style": "smart_casual",
  "color_preferences": ["navy", "gray"],
  "fashion_trendiness": 0.5
}
```

#### 获取 Avatar 上下文

```
GET /api/v1/avatar/{identity_id}/context?scene=meeting&social_context=professional
```

---

### 表情动作 (Expression) - v5.4.0

#### 获取表情画像

```
GET /api/v1/expression/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "facial_expression_settings": {
    "default_expression": "neutral",
    "expressiveness_level": 0.6,
    "eye_contact_frequency": 0.7,
    "smile_frequency": 0.5,
    "natural_blink_rate": 15
  },
  "body_gesture_settings": {
    "gesture_frequency": 0.5,
    "gesture_amplitude": 0.4,
    "posture_style": "confident",
    "hand_movement_style": "natural",
    "head_movement_frequency": 0.3
  }
}
```

#### 更新表情设置

```
PUT /api/v1/expression/{identity_id}/facial
```

**请求体:**

```json
{
  "default_expression": "warm_smile",
  "expressiveness_level": 0.7,
  "eye_contact_frequency": 0.8,
  "smile_frequency": 0.6
}
```

#### 更新手势设置

```
PUT /api/v1/expression/{identity_id}/gesture
```

**请求体:**

```json
{
  "gesture_frequency": 0.6,
  "posture_style": "open",
  "hand_movement_style": "expressive"
}
```

#### 生成表情

```
POST /api/v1/expression/{identity_id}/generate
```

**请求体:**

```json
{
  "emotion": "joy",
  "intensity": 0.8,
  "scene": "meeting"
}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "expression_name": "warm_smile",
  "intensity": 0.7,
  "duration_ms": 500,
  "eyebrow_state": "relaxed",
  "eye_state": "bright",
  "mouth_state": "smile"
}
```

#### 获取表情上下文

```
GET /api/v1/expression/{identity_id}/context?emotion=joy&scene=meeting
```

---

### 表达内容 (Speech) - v5.5.0

#### 获取语音画像

```
GET /api/v1/speech/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "content_style": {
    "formality_level": 0.5,
    "emotional_intensity": 0.6,
    "humor_level": 0.4,
    "directness": 0.6,
    "vocabulary_level": "moderate"
  },
  "expression_depth": {
    "thinking_depth": "moderate",
    "self_disclosure_level": 0.5,
    "abstract_concept_level": 0.6
  }
}
```

#### 更新表达风格

```
PUT /api/v1/speech/{identity_id}/style
```

**请求体:**

```json
{
  "formality_level": 0.6,
  "emotional_intensity": 0.5,
  "humor_level": 0.3,
  "directness": 0.7,
  "vocabulary_level": "professional"
}
```

#### 获取表达上下文

```
GET /api/v1/speech/{identity_id}/context?emotion=joy&scene=meeting
```

---

### TTS 语音合成 (TTS) - v5.6.x

#### 语音合成

```
POST /api/v1/tts/synthesize
```

**请求体:**

```json
{
  "identity_id": "user-123",
  "text": "你好，我是OFA数字人助手！",
  "voice_id": "zh_female_meilinvyou_uranus_bigtts",
  "format": "mp3",
  "sample_rate": 24000,
  "rate": 1.0,
  "pitch": 1.0,
  "volume": 0.7,
  "emotion": "joy",
  "streaming": false
}
```

**响应:**

```json
{
  "session_id": "tts_abc12345",
  "audio_data": "base64_encoded_audio...",
  "duration_ms": 3500,
  "format": "mp3",
  "voice_used": "zh_female_meilinvyou_uranus_bigtts",
  "provider": "doubao",
  "latency_ms": 250,
  "success": true
}
```

#### 获取可用音色列表

```
GET /api/v1/tts/voices?provider=doubao
```

**响应:**

```json
{
  "voices": [
    {
      "voice_id": "zh_male_sunwukong_uranus_bigtts",
      "name": "猴哥",
      "language": "zh-CN",
      "gender": "male",
      "age": "adult",
      "provider": "doubao",
      "description": "孙悟空音色"
    },
    {
      "voice_id": "zh_female_meilinvyou_uranus_bigtts",
      "name": "魅力女友",
      "language": "zh-CN",
      "gender": "female",
      "age": "young",
      "provider": "doubao"
    }
  ]
}
```

#### 声音克隆

```
POST /api/v1/tts/clone
```

**请求体:**

```json
{
  "identity_id": "user-123",
  "voice_name": "我的声音",
  "language": "zh-CN",
  "reference_audios": [
    {
      "audio_url": "https://example.com/voice-sample.mp3",
      "duration_ms": 10000,
      "transcription": "这是参考音频的文本内容"
    }
  ]
}
```

#### 设置身份音色

```
PUT /api/v1/tts/identity/{id}/voice
```

**请求体:**

```json
{
  "voice_id": "zh_female_meilinvyou_uranus_bigtts"
}
```

#### 获取身份音色

```
GET /api/v1/tts/identity/{id}/voice
```

---

## 状态推送

Center 通过 WebSocket 向设备端推送状态更新：

### WebSocket 连接

```
ws://localhost:8080/ws/{identity_id}/{device_id}
```

### 推送消息格式

```json
{
  "type": "state_update",
  "module": "emotion",
  "data": {
    "current_emotion": "joy",
    "intensity": 0.8
  },
  "timestamp": 1711622400000
}
```

### 推送类型

| type | 说明 |
|------|------|
| `state_update` | 状态更新 |
| `decision_context` | 决策上下文 |
| `scene_change` | 场景变化 |
| `sync_request` | 同步请求 |

---

## gRPC API

### 服务列表

| 服务 | 说明 |
|------|------|
| AgentService | Agent 管理与任务执行 |
| MessageService | 消息通信 |
| ManagementService | 系统管理 |
| SoulService | 灵魂特征管理 (v4.x) |
| PresentationService | 外在呈现管理 (v5.x) |

---

## 枚举类型

### EmotionType

| 值 | 说明 |
|---|------|
| JOY | 喜 |
| ANGER | 怒 |
| SADNESS | 哀 |
| FEAR | 惧 |
| LOVE | 爱 |
| DISGUST | 恶 |
| DESIRE | 欲 |

### LifeStage

| 值 | 说明 |
|---|------|
| CHILDHOOD | 童年 |
| ADOLESCENCE | 青春期 |
| EARLY_ADULTHOOD | 青年 |
| YOUNG_ADULTHOOD | 成年早期 |
| MIDDLE_AGE | 中年 |
| MATURE | 成熟期 |
| ELDERLY | 老年 |

### AttachmentStyle

| 值 | 说明 |
|---|------|
| SECURE | 安全型 |
| ANXIOUS | 焦虑型 |
| AVOIDANT | 回避型 |
| DISORGANIZED | 混乱型 |

---

## 错误处理

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

### 错误响应格式

```json
{
  "error": "Invalid parameter: identity_id is required"
}
```

---

## API 端点总览

| 模块 | 端点 | 方法 |
|------|------|------|
| **基础** | `/health` | GET |
| **Identity** | `/api/v1/identities` | POST/GET |
| **Identity** | `/api/v1/identities/{id}` | GET/PUT/DELETE |
| **Device** | `/api/v1/devices` | POST/GET |
| **Device** | `/api/v1/devices/{id}` | GET/PUT/DELETE |
| **Device** | `/api/v1/devices/{id}/heartbeat` | POST |
| **Behavior** | `/api/v1/behaviors` | POST |
| **Behavior** | `/api/v1/behaviors/{identity_id}` | GET |
| **Sync** | `/api/v1/sync` | POST |
| **Sync** | `/api/v1/sync/{identity_id}/state` | GET |
| **Agents** | `/api/v1/agents` | GET |
| **Agents** | `/api/v1/agents/{id}` | GET/DELETE |
| **Tasks** | `/api/v1/tasks` | POST/GET |
| **Tasks** | `/api/v1/tasks/{id}` | GET |
| **Tasks** | `/api/v1/tasks/{id}/cancel` | POST |
| **Messages** | `/api/v1/messages` | POST |
| **Messages** | `/api/v1/messages/broadcast` | POST |
| **Messages** | `/api/v1/messages/multicast` | POST |
| **Skills** | `/api/v1/skills` | GET |
| **System** | `/api/v1/system/info` | GET |
| **System** | `/api/v1/system/metrics` | GET |
| **Emotion** | `/api/v1/emotions/trigger` | POST |
| **Emotion** | `/api/v1/emotions/{identity_id}` | GET |
| **Emotion** | `/api/v1/emotions/{identity_id}/context` | GET |
| **Emotion** | `/api/v1/emotions/{identity_id}/profile` | GET/PUT |
| **Philosophy** | `/api/v1/philosophy/worldview` | POST |
| **Philosophy** | `/api/v1/philosophy/{identity_id}/worldview` | GET |
| **Philosophy** | `/api/v1/philosophy/{identity_id}/context` | GET |
| **Social** | `/api/v1/social/{identity_id}` | GET/PUT |
| **Social** | `/api/v1/social/{identity_id}/education` | GET/PUT |
| **Social** | `/api/v1/social/{identity_id}/career` | GET/PUT |
| **Social** | `/api/v1/social/{identity_id}/context` | GET |
| **Culture** | `/api/v1/culture/{identity_id}` | GET/PUT |
| **Culture** | `/api/v1/culture/{identity_id}/location` | POST |
| **Culture** | `/api/v1/culture/{identity_id}/context` | GET |
| **LifeStage** | `/api/v1/lifestage/{identity_id}` | GET/PUT |
| **LifeStage** | `/api/v1/lifestage/{identity_id}/stage` | POST |
| **LifeStage** | `/api/v1/lifestage/{identity_id}/event` | POST |
| **LifeStage** | `/api/v1/lifestage/{identity_id}/context` | GET |
| **Relationship** | `/api/v1/relationship/{identity_id}` | GET/PUT |
| **Relationship** | `/api/v1/relationship/{identity_id}/add` | POST |
| **Relationship** | `/api/v1/relationship/{identity_id}/context` | GET |
| **Avatar** | `/api/v1/avatar/{identity_id}` | GET/PUT |
| **Avatar** | `/api/v1/avatar/{identity_id}/facial` | PUT |
| **Avatar** | `/api/v1/avatar/{identity_id}/body` | PUT |
| **Avatar** | `/api/v1/avatar/{identity_id}/style` | PUT |
| **Avatar** | `/api/v1/avatar/{identity_id}/context` | GET |
| **Expression** | `/api/v1/expression/{identity_id}` | GET |
| **Expression** | `/api/v1/expression/{identity_id}/facial` | PUT |
| **Expression** | `/api/v1/expression/{identity_id}/gesture` | PUT |
| **Expression** | `/api/v1/expression/{identity_id}/generate` | POST |
| **Expression** | `/api/v1/expression/{identity_id}/context` | GET |
| **Speech** | `/api/v1/speech/{identity_id}` | GET |
| **Speech** | `/api/v1/speech/{identity_id}/style` | PUT |
| **Speech** | `/api/v1/speech/{identity_id}/context` | GET |
| **TTS** | `/api/v1/tts/synthesize` | POST |
| **TTS** | `/api/v1/tts/voices` | GET |
| **TTS** | `/api/v1/tts/clone` | POST |
| **TTS** | `/api/v1/tts/identity/{id}/voice` | GET/PUT |

---

*最后更新: 2026-04-11*
*版本: v6.3.0*