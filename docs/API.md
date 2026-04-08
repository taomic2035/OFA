# OFA API 文档

## 概述

OFA提供两种API接口：
- **REST API**: HTTP接口，端口8080
- **gRPC API**: 高性能RPC接口，端口9090
- **状态推送**: Center 向设备端推送状态更新

---

## 版本概览

| 版本系列 | 特性 | API 命名空间 |
|---------|------|-------------|
| v2.x | 去中心化架构 | `/api/v1/identity`, `/api/v1/sync` |
| v3.x | 多设备协同 | `/api/v1/device`, `/api/v1/message` |
| v4.x | 灵魂特征 | `/api/v1/soul/*` |
| v5.x | 外在呈现 | `/api/v1/presentation/*`, `/api/v1/tts/*` |

---

## REST API

### 基础信息

- 基础URL: `http://localhost:8080`
- 内容类型: `application/json`
- 认证: 无 (开发版本)

---

## v4.x 灵魂特征 API

### 情绪系统 (v4.0.0)

#### 获取情绪状态

```
GET /api/v1/soul/emotion/{identity_id}
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
  "emotion_stability": 0.7,
  "last_trigger": "positive_interaction",
  "timestamp": 1711622400000
}
```

#### 更新情绪状态

```
POST /api/v1/soul/emotion/{identity_id}/update
```

**请求体:**

```json
{
  "trigger_type": "event",
  "trigger_name": "received_compliment",
  "intensity": 0.8,
  "context": {
    "source": "user_interaction",
    "description": "Received a compliment from friend"
  }
}
```

#### 获取欲望状态

```
GET /api/v1/soul/desire/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "desires": {
    "physiological": {
      "hunger": 0.3,
      "thirst": 0.2,
      "sleep": 0.4
    },
    "safety": {
      "security": 0.8,
      "stability": 0.7
    },
    "social": {
      "belonging": 0.6,
      "intimacy": 0.5
    },
    "esteem": {
      "recognition": 0.5,
      "achievement": 0.4
    },
    "self_actualization": {
      "growth": 0.6,
      "creativity": 0.7
    }
  },
  "dominant_desire": "creativity",
  "satisfaction_level": 0.65
}
```

---

### 三观系统 (v4.1.0)

#### 获取世界观

```
GET /api/v1/soul/worldview/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "world_essence": "material",
  "world_order": "orderly",
  "human_nature": "mixed",
  "social_cognition": {
    "society_fairness": 0.6,
    "social_mobility": 0.5,
    "institutional_trust": 0.7
  },
  "future_view": {
    "optimism": 0.7,
    "change_orientation": "progressive",
    "control_belief": 0.6
  }
}
```

#### 获取人生观

```
GET /api/v1/soul/lifeview/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "meaning_source": "contribution",
  "time_orientation": "present_future",
  "life_attitude": {
    "adventure_seeking": 0.6,
    "risk_tolerance": 0.5,
    "control_focus": "internal"
  },
  "success_definition": "balance",
  "happiness_source": ["relationships", "growth", "health"]
}
```

#### 获取价值观

```
GET /api/v1/soul/values/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "core_values": {
    "honesty": 0.9,
    "kindness": 0.85,
    "family": 0.8,
    "health": 0.75,
    "freedom": 0.7,
    "achievement": 0.65,
    "creativity": 0.6
  },
  "moral_framework": "utilitarian",
  "decision_priority": ["honesty", "kindness", "family"]
}
```

---

### 社会身份 (v4.2.0)

#### 获取社会身份画像

```
GET /api/v1/soul/identity/{identity_id}
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
  },
  "identity_profile": {
    "primary_role": "professional",
    "role_priority": ["professional", "family_member", "friend"],
    "identity_confidence": 0.75
  }
}
```

---

### 地域文化 (v4.3.0)

#### 获取地域文化画像

```
GET /api/v1/soul/culture/{identity_id}
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
  "social_style": "open",
  "migration_history": []
}
```

---

### 人生阶段 (v4.4.0)

#### 获取人生阶段状态

```
GET /api/v1/soul/lifestage/{identity_id}
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

---

### 情绪行为联动 (v4.5.0)

#### 获取情绪行为状态

```
GET /api/v1/soul/behavior/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "decision_influence": {
    "risk_preference": "moderate",
    "impulse_control": 0.7,
    "social_tendency": 0.6,
    "decision_style": "analytical"
  },
  "expression_influence": {
    "tone_style": "warm",
    "word_choice": "positive",
    "expression_frequency": 0.6
  },
  "triggered_behaviors": [
    {
      "behavior_type": "seek_social_connection",
      "action_tendency": "approach",
      "urgency": 0.4
    }
  ],
  "coping_strategies": ["problem_focused", "seeking_support"]
}
```

---

### 人际关系 (v4.6.0)

#### 获取人际关系状态

```
GET /api/v1/soul/relationship/{identity_id}
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

---

## v5.x 外在呈现 API

### 外在形象系统 (v5.0.0)

#### 获取 Avatar 形象

```
GET /api/v1/presentation/avatar/{identity_id}
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
    "height": 175,
    "weight": 70,
    "body_type": "average",
    "posture": "confident",
    "movement_style": "energetic"
  },
  "age_appearance": {
    "apparent_age": 28,
    "aging_stage": "prime",
    "facial_maturity": 0.6
  },
  "style_preferences": {
    "clothing_style": "casual",
    "accessory_style": "minimal",
    "overall_vibe": "professional"
  },
  "model_3d": {
    "model_id": "avatar-001",
    "model_type": "custom",
    "render_quality": "high"
  }
}
```

#### 更新 Avatar 形象

```
PUT /api/v1/presentation/avatar/{identity_id}
```

**请求体:**

```json
{
  "facial_features": {
    "hair_style": "medium",
    "hair_color": "brown"
  },
  "style_preferences": {
    "clothing_style": "smart_casual"
  }
}
```

---

### 语音合成系统 (v5.1.0 - v5.6.x)

#### 获取语音配置

```
GET /api/v1/presentation/voice/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "characteristics": {
    "voice_type": "male_young",
    "pitch": 0.5,
    "speed": 1.0,
    "volume": 0.7,
    "timbre": "warm"
  },
  "style": {
    "formality": "casual",
    "emotion_expressiveness": 0.6,
    "accent_style": "neutral"
  },
  "emotional_voice": {
    "joy_voice": {"pitch_modifier": 0.1, "speed_modifier": 0.1},
    "anger_voice": {"pitch_modifier": -0.1, "speed_modifier": 0.2}
  },
  "tts_config": {
    "engine": "standard",
    "quality": "high",
    "streaming": true
  }
}
```

#### 生成语音

```
POST /api/v1/presentation/voice/{identity_id}/synthesize
```

**请求体:**

```json
{
  "text": "你好，很高兴见到你！",
  "emotion": "joy",
  "speed": 1.0,
  "output_format": "mp3"
}
```

**响应:**

```json
{
  "audio_url": "/audio/output-001.mp3",
  "duration_ms": 2500,
  "format": "mp3",
  "sample_rate": 22050
}
```

---

### TTS引擎API (v5.6.x)

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

**可用音色类型:**
- 大模型音色: 猴哥、魅力女友、儒雅逸辰、爽朗少年等30+音色
- 情绪音色: 开心女声、悲伤女声、愤怒女声
- 方言音色: 东北方言、四川方言、广东方言

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

**响应:**

```json
{
  "voice_id": "cloned_voice_user-123",
  "voice_name": "我的声音",
  "status": "ready",
  "quality": 0.85,
  "message": "Voice cloning completed successfully"
}
```

#### 设置身份音色映射

```
PUT /api/v1/tts/identity/{identity_id}/voice
```

**请求体:**

```json
{
  "voice_id": "zh_female_meilinvyou_uranus_bigtts"
}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "voice_id": "zh_female_meilinvyou_uranus_bigtts",
  "success": true
}
```

#### 获取身份音色

```
GET /api/v1/tts/identity/{identity_id}/voice
```

**响应:**

```json
{
  "identity_id": "user-123",
  "voice_id": "zh_female_meilinvyou_uranus_bigtts"
}
```

---

### 表达内容系统 (v5.2.0)

#### 获取表达内容配置

```
GET /api/v1/presentation/speech/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "content_style": {
    "tone_style": "warm",
    "language_level": "moderate",
    "directness": 0.6,
    "humor_tendency": 0.4,
    "emotional_coloring": "positive"
  },
  "expression_depth": {
    "thinking_depth": "moderate",
    "self_disclosure_level": 0.5,
    "abstract_concept_level": 0.6
  },
  "cultural_expression": {
    "indirect_expression": 0.4,
    "respect_level": 0.7,
    "honorific_usage": "moderate",
    "face_awareness": 0.5
  },
  "social_expression": {
    "professional_tone": "balanced",
    "authority_expression": "confident",
    "identity_confidence": 0.75
  }
}
```

#### 获取决策上下文

```
GET /api/v1/presentation/speech/{identity_id}/context
```

**响应:**

```json
{
  "identity_id": "user-123",
  "recommended_tone": "warm_professional",
  "recommended_formality": "moderate",
  "recommended_depth": "moderate",
  "recommended_length": "medium",
  "scene_adaptation": {
    "scene": "meeting",
    "formality_level": 0.7,
    "expression_range": "professional"
  },
  "emotion_adaptation": {
    "current_emotion": "joy",
    "expression_intensity": 0.6
  },
  "opening_suggestion": "您好，有什么可以帮您的吗？",
  "closing_suggestion": "祝您有愉快的一天！"
}
```

---

### 表情动作系统 (v5.3.0)

#### 获取表情动作配置

```
GET /api/v1/presentation/expression/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "facial_expression_settings": {
    "default_expression": "neutral",
    "expression_intensity": 0.6,
    "eye_contact_tendency": 0.7,
    "blink_rate": 15.0,
    "smile_tendency": 0.5,
    "micro_expression_enabled": true
  },
  "body_gesture_settings": {
    "default_posture": "confident",
    "gesture_intensity": 0.5,
    "hand_gesture_enabled": true,
    "nod_frequency": 0.4,
    "mirroring_enabled": false
  },
  "emotion_expression_mapping": {
    "joy": {
      "expression_type": "smile",
      "intensity": 0.8,
      "posture": "open"
    },
    "anger": {
      "expression_type": "furrowed_brow",
      "intensity": 0.6,
      "posture": "tense"
    }
  },
  "animation_settings": {
    "lip_sync_enabled": true,
    "breathing_animation_enabled": true,
    "eye_movement_enabled": true
  }
}
```

#### 获取当前表情状态

```
GET /api/v1/presentation/expression/{identity_id}/current
```

**响应:**

```json
{
  "identity_id": "user-123",
  "current_expression": {
    "expression_name": "warm_smile",
    "intensity": 0.7,
    "duration_ms": 500,
    "eyebrow_state": "relaxed",
    "eye_state": "bright",
    "mouth_state": "smile"
  },
  "current_gesture": {
    "gesture_name": "open_posture",
    "posture": "confident",
    "hand_position": "natural"
  },
  "recommended_expression": {
    "expression_name": "attentive",
    "reason": "listening_mode"
  }
}
```

---

### 形象个性化系统 (v5.4.0)

#### 获取个性化配置

```
GET /api/v1/presentation/personalization/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "image_preferences": {
    "preferred_colors": ["blue", "black", "white"],
    "preferred_styles": ["casual", "smart_casual"],
    "style_experimentation": 0.4,
    "comfort_priority": "medium",
    "presentation_effort": "medium"
  },
  "image_evolution": {
    "evolution_mode": "gradual",
    "seasonal_adaptation": true,
    "trend_following_level": 0.4,
    "core_style_elements": ["minimalist", "clean_lines"]
  },
  "scene_adaptation_settings": {
    "adaptation_mode": "auto",
    "default_work_style": "business_casual",
    "default_home_style": "relaxed",
    "location_awareness": true
  },
  "style_management": {
    "recommendation_enabled": true,
    "recommendation_source": "hybrid"
  }
}
```

#### 获取个性化上下文

```
GET /api/v1/presentation/personalization/{identity_id}/context
```

**响应:**

```json
{
  "identity_id": "user-123",
  "recommended_style": {
    "style_name": "smart_casual",
    "confidence": 0.75,
    "color_palette": ["navy", "white", "gray"],
    "occasion": "work",
    "season": "spring"
  },
  "style_score": 0.72,
  "consistency_score": 0.68,
  "versatility_score": 0.65,
  "evolution_readiness": 0.55,
  "style_suggestions": [
    {
      "suggestion_type": "add",
      "target_area": "accessory",
      "suggestion": "Consider adding a watch for professional settings"
    }
  ]
}
```

---

### 多端展示系统 (v5.5.0)

#### 获取展示配置

```
GET /api/v1/presentation/display/{identity_id}
```

**响应:**

```json
{
  "identity_id": "user-123",
  "rendering_settings": {
    "render_engine": "webgl",
    "quality_preset": "high",
    "target_fps": 60,
    "adaptive_quality": true,
    "post_processing_enabled": true,
    "physics_enabled": true
  },
  "device_adaptation": {
    "adaptation_mode": "auto",
    "auto_optimize": true,
    "mobile_optimizations": {
      "gpu_power_mode": "balanced",
      "texture_streaming": true,
      "battery_aware_mode": true
    }
  },
  "display_sync": {
    "sync_mode": "state_sync",
    "sync_frequency_ms": 100,
    "interpolation_mode": "linear",
    "prediction_enabled": true
  }
}
```

#### 获取展示上下文

```
GET /api/v1/presentation/display/{identity_id}/context
```

**响应:**

```json
{
  "identity_id": "user-123",
  "current_device": "phone-primary",
  "device_type": "phone",
  "current_quality": "high",
  "current_fps": 58.5,
  "sync_status": "synced",
  "sync_latency_ms": 45,
  "battery_level": 0.75,
  "thermal_state": "normal",
  "connected_devices": ["watch-001", "tablet-001"],
  "recommended_quality": "high",
  "quality_adjustment_needed": false
}
```

#### 同步展示状态

```
POST /api/v1/presentation/display/{identity_id}/sync
```

**请求体:**

```json
{
  "device_id": "phone-primary",
  "state": {
    "avatar_position": {"x": 0, "y": 0, "z": 0},
    "avatar_rotation": {"x": 0, "y": 0, "z": 0},
    "current_animation": "idle",
    "current_expression": "neutral"
  },
  "timestamp": 1711622400000
}
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

## 基础 REST API

### 健康检查

```
GET /health
```

**响应:**

```json
{
  "status": "healthy",
  "version": "v5.5.0"
}
```

---

### Agent 管理

#### 获取Agent列表

```
GET /api/v1/agents
```

**查询参数:**

| 参数 | 类型 | 说明 |
|------|------|------|
| type | int | Agent类型过滤 |
| status | int | 状态过滤 |
| page | int | 页码，默认1 |
| page_size | int | 每页数量，默认20 |

#### 获取单个Agent

```
GET /api/v1/agents/{id}
```

#### 删除Agent

```
DELETE /api/v1/agents/{id}
```

---

### 任务管理

#### 提交任务

```
POST /api/v1/tasks
```

**请求体:**

```json
{
  "skill_id": "text.process",
  "input": "eyJ0ZXh0IjoiaGVsbG8iLCJvcGVyYXRpb24iOiJ1cHBlcmNhc2UifQ==",
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

### 消息通信

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

## gRPC API

### 服务列表

| 服务 | 说明 |
|------|------|
| AgentService | Agent管理与任务执行 |
| MessageService | 消息通信 |
| ManagementService | 系统管理 |
| SoulService | 灵魂特征管理 (v4.x) |
| PresentationService | 外在呈现管理 (v5.x) |

### AgentService

```protobuf
service AgentService {
  rpc Connect(stream AgentMessage) returns (stream CenterMessage);
  rpc SubmitTask(SubmitTaskRequest) returns (SubmitTaskResponse);
  rpc GetTaskStatus(GetTaskStatusRequest) returns (GetTaskStatusResponse);
  rpc CancelTask(CancelTaskRequest) returns (CancelTaskResponse);
  rpc SubscribeTask(SubscribeTaskRequest) returns (stream TaskEvent);
  rpc RegisterCapabilities(RegisterCapabilitiesRequest) returns (RegisterCapabilitiesResponse);
  rpc GetCapabilities(GetCapabilitiesRequest) returns (GetCapabilitiesResponse);
}
```

### SoulService (v4.x)

```protobuf
service SoulService {
  // 情绪系统
  rpc GetEmotionState(GetEmotionStateRequest) returns (EmotionState);
  rpc UpdateEmotion(UpdateEmotionRequest) returns (EmotionState);
  rpc GetDesireState(GetDesireStateRequest) returns (DesireState);

  // 三观系统
  rpc GetWorldview(GetWorldviewRequest) returns (Worldview);
  rpc GetLifeView(GetLifeViewRequest) returns (LifeView);
  rpc GetValueSystem(GetValueSystemRequest) returns (ValueSystem);

  // 社会身份
  rpc GetSocialIdentity(GetSocialIdentityRequest) returns (SocialIdentity);

  // 地域文化
  rpc GetRegionalCulture(GetRegionalCultureRequest) returns (RegionalCulture);

  // 人生阶段
  rpc GetLifeStage(GetLifeStageRequest) returns (LifeStage);

  // 情绪行为
  rpc GetEmotionBehavior(GetEmotionBehaviorRequest) returns (EmotionBehavior);

  // 人际关系
  rpc GetRelationship(GetRelationshipRequest) returns (RelationshipState);
}
```

### PresentationService (v5.x)

```protobuf
service PresentationService {
  // 外在形象
  rpc GetAvatar(GetAvatarRequest) returns (Avatar);
  rpc UpdateAvatar(UpdateAvatarRequest) returns (Avatar);

  // 语音合成
  rpc GetVoiceProfile(GetVoiceProfileRequest) returns (VoiceProfile);
  rpc SynthesizeVoice(SynthesizeVoiceRequest) returns (SynthesizeVoiceResponse);

  // 表达内容
  rpc GetSpeechContent(GetSpeechContentRequest) returns (SpeechContentProfile);
  rpc GetSpeechContext(GetSpeechContextRequest) returns (SpeechDecisionContext);

  // 表情动作
  rpc GetExpressionGesture(GetExpressionGestureRequest) returns (ExpressionGestureProfile);
  rpc GetCurrentExpression(GetCurrentExpressionRequest) returns (ExpressionGestureContext);

  // 形象个性化
  rpc GetPersonalization(GetPersonalizationRequest) returns (AvatarPersonalizationProfile);
  rpc GetPersonalizationContext(GetPersonalizationContextRequest) returns (PersonalizationContext);

  // 多端展示
  rpc GetDisplaySettings(GetDisplaySettingsRequest) returns (MultiDisplayProfile);
  rpc GetDisplayContext(GetDisplayContextRequest) returns (DisplayContext);
  rpc SyncDisplayState(SyncDisplayStateRequest) returns (SyncDisplayStateResponse);
}
```

---

## 枚举类型

### AgentType

| 值 | 说明 |
|---|------|
| 0 | UNKNOWN |
| 1 | FULL (完整版) |
| 2 | MOBILE (移动版) |
| 3 | LITE (轻量版) |
| 4 | IOT (物联网) |
| 5 | EDGE (边缘计算) |

### AgentStatus

| 值 | 说明 |
|---|------|
| 0 | UNKNOWN |
| 1 | ONLINE |
| 2 | BUSY |
| 3 | IDLE |
| 4 | OFFLINE |

### TaskStatus

| 值 | 说明 |
|---|------|
| 0 | UNKNOWN |
| 1 | PENDING |
| 2 | RUNNING |
| 3 | COMPLETED |
| 4 | FAILED |
| 5 | CANCELLED |
| 6 | TIMEOUT |

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

### QualityPreset

| 值 | 说明 |
|---|------|
| LOW | 低质量 |
| MEDIUM | 中等质量 |
| HIGH | 高质量 |
| ULTRA | 超高质量 |
| CUSTOM | 自定义 |

---

## 错误处理

### HTTP状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

### 错误响应格式

```json
{
  "error": {
    "code": "INVALID_PARAMETER",
    "message": "Invalid parameter: identity_id is required",
    "details": {}
  }
}
```

---

*最后更新: 2026-04-08*
*版本: v5.6.5*