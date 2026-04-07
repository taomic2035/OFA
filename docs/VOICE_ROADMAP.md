# OFA 语音合成系统路线图

## 概述

语音合成是数字人"外在呈现"的核心组件，目标是实现 **音色复刻 + 语调一致 + 风格匹配 + 灵魂同步** 的完整语音能力。

**核心理念**: 语音是灵魂的外在表达，需要与情绪、三观、文化、身份等内在特征保持一致。

---

## 当前状态 (v5.1.0)

### 已实现

| 模块 | 状态 | 说明 |
|------|------|------|
| VoiceCharacteristics | ✅ | 音高/语速/音量/音色参数模型 |
| VoiceStyle | ✅ | 口音/正式度/沟通风格 |
| EmotionalVoice | ✅ | 七情(喜怒哀惧爱恶欲)语音映射 |
| SpeechPatterns | ✅ | 词汇级别/句式/停顿模式 |
| TTSConfiguration | ✅ | 引擎配置框架 |
| 情绪影响 | ✅ | Emotion → 语音参数映射 |
| 文化影响 | ✅ | RegionalCulture → 口音/风格 |
| 人生阶段影响 | ✅ | LifeStage → 声音年龄 |
| 场景适应 | ✅ | 场景 → 语速/音量/正式度 |
| 社交适应 | ✅ | 社交语境 → 权威/温暖/直接度 |

### 未实现

| 能力 | 状态 | 优先级 |
|------|------|--------|
| 音色复刻 | ❌ | P0 |
| TTS引擎集成 | ❌ | P0 |
| 语音特征提取 | ❌ | P0 |
| 韵律精细控制 | ❌ | P1 |
| 多语言/方言支持 | ❌ | P1 |
| 实时流式合成 | ❌ | P1 |
| 语音情感迁移 | ❌ | P2 |

---

## 版本规划

### v6.0.0 - 音色复刻与TTS引擎集成

**目标**: 实现真正的语音合成能力，支持音色复刻

#### 1. VoiceCloningModel (音色复刻模型)

```go
type VoiceCloningModel struct {
    IdentityID       string    `json:"identity_id"`

    // 参考音频
    ReferenceAudios  []ReferenceAudio `json:"reference_audios"`

    // 声纹特征
    VoiceFingerprint VoiceFingerprint `json:"voice_fingerprint"`

    // 复刻模型
    ClonedModelID    string    `json:"cloned_model_id"`
    ClonedModelURL   string    `json:"cloned_model_url"`
    ModelProvider    string    `json:"model_provider"` // elevenlabs, azure, custom

    // 复刻状态
    CloneStatus      string    `json:"clone_status"` // pending, processing, ready, failed
    CloneQuality     double    `json:"clone_quality"` // 0-1
    CloneTimestamp   time.Time `json:"clone_timestamp"`
}

type ReferenceAudio struct {
    AudioID       string  `json:"audio_id"`
    AudioURL      string  `json:"audio_url"`
    DurationMs    int     `json:"duration_ms"`
    Quality       double  `json:"quality"`
    Transcription string  `json:"transcription"` // 文本对齐
    Emotion       string  `json:"emotion"`       // 情绪标签
    Language      string  `json:"language"`
}

type VoiceFingerprint struct {
    // 声学特征
    MFCC           []float64 `json:"mfcc"`           // 梅尔频率倒谱系数
    PitchMean      double    `json:"pitch_mean"`     // 平均音高
    PitchStd       double    `json:"pitch_std"`      // 音高标准差
    SpectralTilt   double    `json:"spectral_tilt"`  // 频谱倾斜
    Formants       []double  `json:"formants"`       // 共振峰 F1-F4

    // 韵律特征
    SpeechRate     double    `json:"speech_rate"`    // 平均语速
    PausePattern   []double  `json:"pause_pattern"`  // 停顿模式
    IntensityVar   double    `json:"intensity_var"`  // 强度变化

    // 音色特征
    TimbreVector   []float64 `json:"timbre_vector"`  // 音色嵌入向量
    Breathiness    double    `json:"breathiness"`
    Roughness      double    `json:"roughness"`
}
```

#### 2. TTSEngine (TTS引擎集成)

```go
type TTSEngine struct {
    // 多引擎支持
    PrimaryEngine   TTSProvider
    FallbackEngine  TTSProvider

    // 引擎池
    EnginePool      map[string]TTSProvider // elevenlabs, azure, google, custom
}

type TTSProvider interface {
    // 基础合成
    Synthesize(ctx context.Context, req *SynthesisRequest) (*SynthesisResult, error)

    // 流式合成
    SynthesizeStream(ctx context.Context, req *SynthesisRequest) (<-chan AudioChunk, error)

    // 音色复刻
    CloneVoice(ctx context.Context, req *CloneRequest) (*CloneResult, error)

    // 声音管理
    ListVoices(ctx context.Context) ([]VoiceInfo, error)
    GetVoice(ctx context.Context, voiceID string) (*VoiceInfo, error)

    // 能力查询
    SupportsStreaming() bool
    SupportsCloning() bool
    SupportsSSML() bool
    SupportedLanguages() []string
}

type SynthesisRequest struct {
    Text         string            `json:"text"`
    VoiceID      string            `json:"voice_id"`
    Language     string            `json:"language"`

    // 语音参数
    Pitch        double            `json:"pitch"`
    Rate         double            `json:"rate"`
    Volume       double            `json:"volume"`

    // 情感/风格
    Emotion      string            `json:"emotion"`
    Style        string            `json:"style"`

    // 输出设置
    OutputFormat string            `json:"output_format"`
    SampleRate   int               `json:"sample_rate"`

    // SSML
    SSML         string            `json:"ssml,omitempty"`

    // 流式
    Streaming    bool              `json:"streaming"`
}

type SynthesisResult struct {
    AudioData    []byte  `json:"audio_data"`
    AudioURL     string  `json:"audio_url"`
    DurationMs   int     `json:"duration_ms"`
    Format       string  `json:"format"`
    SampleRate   int     `json:"sample_rate"`

    // 质量指标
    LatencyMs    int     `json:"latency_ms"`
    QualityScore double  `json:"quality_score"`
}
```

#### 3. 引擎适配器

```
src/center/internal/tts/
├── engine.go              # TTS引擎核心
├── provider.go            # Provider接口
├── providers/
│   ├── elevenlabs.go      # ElevenLabs API
│   ├── azure.go           # Azure Speech Services
│   ├── google.go          # Google Cloud TTS
│   ├── alibaba.go         # 阿里云语音合成
│   ├── local.go           # 本地VITS/SoVITS
│   └── custom.go          # 自定义模型服务
├── cloning/
│   ├── voice_cloner.go    # 音色复刻服务
│   ├── feature_extractor.go # 特征提取
│   └── fingerprint.go     # 声纹生成
└── cache/
    └── synthesis_cache.go # 合成结果缓存
```

#### 4. 音色复刻流程

```
┌─────────────────────────────────────────────────────────────┐
│                     音色复刻流程                              │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. 音频采集                                                 │
│     用户录音/上传音频 → 多段参考音频 (建议3-10分钟)            │
│                          ↓                                   │
│  2. 特征提取                                                 │
│     音频预处理 → MFCC/共振峰/韵律分析 → VoiceFingerprint      │
│                          ↓                                   │
│  3. 模型训练                                                 │
│     选择Provider → 上传参考音频 → 训练克隆模型                 │
│                          ↓                                   │
│  4. 验证与调优                                               │
│     测试合成 → 相似度评估 → 微调参数                          │
│                          ↓                                   │
│  5. 部署使用                                                 │
│     模型ID绑定 → VoiceProfile关联 → 实时合成                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### v6.1.0 - 韵律精细控制

**目标**: 实现精细的韵律控制，让语音更自然

#### 1. ProsodyModel (韵律模型)

```go
type ProsodyModel struct {
    IdentityID string `json:"identity_id"`

    // 音高曲线
    PitchCurve PitchCurveModel `json:"pitch_curve"`

    // 时长控制
    DurationModel DurationModel `json:"duration_model"`

    // 重音模式
    StressPattern StressPattern `json:"stress_pattern"`

    // 停顿策略
    PauseStrategy PauseStrategy `json:"pause_strategy"`

    // 语调模式
    IntonationPattern IntonationPattern `json:"intonation_pattern"`
}

type PitchCurveModel struct {
    // 句首/句中/句尾音高变化
    SentenceInitial double `json:"sentence_initial"`
    SentenceMedial  double `json:"sentence_medial"`
    SentenceFinal   double `json:"sentence_final"`

    // 陈述句/疑问句/感叹句
    DeclarativeEnd  double `json:"declarative_end"`  // 陈述句结尾
    InterrogativeEnd double `json:"interrogative_end"` // 疑问句结尾
    ExclamatoryEnd  double `json:"exclamatory_end"`   // 感叹句结尾

    // 焦点音高提升
    FocusBoost      double `json:"focus_boost"`
}

type PauseStrategy struct {
    // 句间停顿
    SentencePause   int `json:"sentence_pause"`   // ms
    ClausePause     int `json:"clause_pause"`     // ms
    PhrasePause     int `json:"phrase_pause"`     // ms

    // 情绪停顿
    ThoughtfulPause int `json:"thoughtful_pause"` // 思考停顿
    EmphasisPause   int `json:"emphasis_pause"`   // 强调停顿

    // 个性化
    HesitationRate  double `json:"hesitation_rate"` // 犹豫频率
    BreathPauseRate double `json:"breath_pause_rate"` // 呼吸停顿
}
```

#### 2. SSML生成器

```go
type SSMLGenerator struct {
    prosody *ProsodyModel
}

func (g *SSMLGenerator) Generate(text string, emotion string, context SpeechContext) string {
    // 根据文本分析生成SSML
    // 包含: prosody, emphasis, break, say-as等标签
}
```

---

### v6.2.0 - 多语言与方言支持

**目标**: 支持多语言、方言的语音合成

#### 1. MultiLanguageModel

```go
type MultiLanguageModel struct {
    IdentityID string `json:"identity_id"`

    // 主语言
    PrimaryLanguage string `json:"primary_language"`

    // 多语言能力
    Languages map[string]LanguageProficiency `json:"languages"`

    // 方言能力
    Dialects map[string]DialectProficiency `json:"dialects"`

    // 口音设置
    AccentSettings AccentSettings `json:"accent_settings"`
}

type LanguageProficiency struct {
    Language    string `json:"language"`
    Level       string `json:"level"`       // native, fluent, conversational, basic
    Accent      string `json:"accent"`      // native, slight, moderate, heavy
    Preference  double `json:"preference"`  // 使用偏好 0-1
}

type DialectProficiency struct {
    Dialect     string `json:"dialect"`      // 粤语、四川话、东北话等
    Level       string `json:"level"`
    ActiveUsage bool   `json:"active_usage"` // 是否主动使用
    Intensity   double `json:"intensity"`    // 方言强度 0-1
}

type AccentSettings struct {
    // 外语口音
    ForeignAccentIntensity double `json:"foreign_accent_intensity"`

    // 地域口音
    RegionalAccentIntensity double `json:"regional_accent_intensity"`

    // 正式场合口音调整
    FormalAccentReduction double `json:"formal_accent_reduction"`
}
```

#### 2. 方言映射

```
Region → Dialect
├── Guangdong → Cantonese (粤语)
├── Sichuan → Sichuanese (四川话)
├── Northeast → Northeastern (东北话)
├── Shanghai → Shanghainese (上海话)
├── Fujian → Hokkien (闽南语)
└── ...
```

---

### v6.3.0 - 实时流式与低延迟

**目标**: 实现实时对话场景的低延迟语音合成

#### 1. 流式合成架构

```
┌─────────────────────────────────────────────────────────────┐
│                    流式语音合成流程                           │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  文本流输入 → 分句 → 并行合成 → 音频流输出                     │
│     │          │          │           │                      │
│     │          │          │           ↓                      │
│     │          │          │    AudioChunk[]                  │
│     │          │          │    (50-100ms chunks)             │
│     │          │          │                                   │
│     │          │          └─── 多引擎并行                     │
│     │          │               (主引擎+备引擎)                 │
│     │          │                                              │
│     │          └── 智能分句                                   │
│     │              (按标点/语义/长度)                          │
│     │                                                         │
│     └── 文本流缓冲                                            │
│         (支持打字机效果)                                       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

#### 2. 预测性合成

```go
type PredictiveSynthesizer struct {
    // 预测缓存
    predictionCache *PredictionCache

    // 文本预测模型
    textPredictor TextPredictor
}

// 根据上下文预测可能的说出内容，提前合成缓存
func (s *PredictiveSynthesizer) PredictAndSynthesize(context DialogContext) {
    predictions := s.textPredictor.Predict(context)
    for _, pred := range predictions {
        go s.synthesizeAsync(pred)
    }
}
```

---

## 与灵魂系统的集成

### v4.x 灵魂 → v6.x 语音映射

| 灵魂组件 | 语音影响 | 映射方式 |
|----------|----------|----------|
| **Emotion (v4.0)** | 音高/语速/音色 | EmotionMapping → PitchShift/RateMultiplier |
| **Worldview (v4.1)** | 表达风格 | 世界观 → 语调风格(权威/温和) |
| **LifeView (v4.1)** | 语言模式 | 人生观 → 词汇级别/句式复杂度 |
| **ValueSystem (v4.1)** | 语调特征 | 核心价值观 → 语音温度/直接度 |
| **SocialIdentity (v4.2)** | 正式度 | 社会身份 → 正式度/权威感 |
| **RegionalCulture (v4.3)** | 口音/方言 | 地域 → 口音强度/方言使用 |
| **LifeStage (v4.4)** | 声音年龄 | 人生阶段 → VoiceAge (young/adult/senior) |
| **EmotionBehavior (v4.5)** | 表达强度 | 行为倾向 → 情绪表达强度 |
| **Relationship (v4.6)** | 社交语调 | 关系类型 → 温暖度/距离感 |

### 集成示例

```go
// Center端：综合所有灵魂特征计算最终语音参数
func (e *VoiceEngine) CalculateFinalVoice(identity *PersonalIdentity,
    emotion *Emotion, culture *RegionalCulture, lifeStage *LifeStage,
    scene string, socialContext string) *FinalVoiceParameters {

    params := &FinalVoiceParameters{
        // 基础特征
        VoiceType:  identity.VoiceProfile.VoiceType,
        VoiceAge:   e.calculateVoiceAge(lifeStage),

        // 音高 = 基础 + 情绪影响 + 文化影响
        Pitch: e.calculatePitch(identity, emotion, culture),

        // 语速 = 基础 + 情绪影响 + 场景影响
        Rate: e.calculateRate(identity, emotion, scene),

        // 音色 = 复刻音色 + 情绪调制
        VoiceID: identity.VoiceProfile.ClonedModelID,
        TimbreMod: e.calculateTimbreMod(emotion),

        // 风格 = 文化 + 身份 + 社交
        Style: e.calculateStyle(culture, socialContext),
    }

    return params
}
```

---

## 实现优先级

| Phase | 版本 | 内容 | 工作量 | 依赖 |
|-------|------|------|--------|------|
| **P0** | v6.0.0 | 音色复刻 + TTS引擎集成 | 高 | 外部API |
| **P0** | v6.0.1 | 特征提取 + 声纹生成 | 中 | 音频处理库 |
| **P1** | v6.1.0 | 韵律精细控制 | 中 | v6.0.0 |
| **P1** | v6.2.0 | 多语言/方言支持 | 中 | 多语言TTS |
| **P1** | v6.3.0 | 实时流式合成 | 中 | WebSocket |
| **P2** | v6.4.0 | 语音情感迁移 | 高 | 深度学习 |

---

## 技术选型建议

### TTS引擎选择

| 引擎 | 音色复刻 | 流式 | 质量 | 成本 | 推荐 |
|------|----------|------|------|------|------|
| ElevenLabs | ✅ 优秀 | ✅ | 极高 | 高 | 高端场景 |
| Azure Speech | ✅ 良好 | ✅ | 高 | 中 | 企业场景 |
| Google Cloud TTS | ⚠️ 一般 | ✅ | 高 | 中 | 通用场景 |
| 阿里云语音 | ✅ 良好 | ✅ | 高 | 低 | 中文场景 |
| VITS/SoVITS | ✅ 良好 | ✅ | 中 | 免费 | 本地部署 |

### 推荐架构

```
                    ┌──────────────────┐
                    │   VoiceEngine    │
                    │   (Center)       │
                    └────────┬─────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
        ┌─────▼─────┐  ┌─────▼─────┐  ┌─────▼─────┐
        │  Cloud    │  │   Local   │  │  Hybrid   │
        │  TTS      │  │   TTS     │  │  Router   │
        └───────────┘  └───────────┘  └───────────┘
              │              │
    ┌─────────┼─────────┐    │
    │         │         │    │
ElevenLabs Azure   Alibaba  VITS
(高质量) (稳定)  (中文)   (本地)
```

---

## 下一步行动

1. **v6.0.0 实现**
   - [ ] 实现 TTSEngine 核心框架
   - [ ] 集成 ElevenLabs API (P0)
   - [ ] 集成 Azure Speech Services (P0)
   - [ ] 实现 VoiceCloningModel (P0)
   - [ ] 实现语音特征提取 (P0)

2. **Android SDK 更新**
   - [ ] 添加 TTSCClient 实时合成客户端
   - [ ] 添加 VoiceCloningUI 音色复刻流程
   - [ ] 实现流式音频播放

3. **测试验证**
   - [ ] 音色复刻相似度测试 (>90%)
   - [ ] 实时合成延迟测试 (<200ms首包)
   - [ ] 情绪映射准确性测试

---

*最后更新: 2026-04-07*
*版本: v6.0.0-rc*