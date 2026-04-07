package com.ofa.agent.avatar;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.List;

/**
 * Avatar状态模型 (v5.0.0)
 *
 * 端侧接收 Center 推送的 Avatar 状态，用于调整外在呈现。
 * 深层 Avatar 管理在 Center 端 AvatarEngine 完成。
 */
public class AvatarState {

    // === 面部特征 ===
    private FacialFeatures facialFeatures;

    // === 体型特征 ===
    private BodyFeatures bodyFeatures;

    // === 年龄外观 ===
    private AgeAppearance ageAppearance;

    // === 风格偏好 ===
    private StylePreferences stylePreferences;

    // === 3D模型引用 ===
    private Model3DReference model3D;

    // === 展示设置 ===
    private DisplaySettings displaySettings;

    // === 决策上下文 ===
    private AvatarDecisionContext decisionContext;

    // === 版本信息 ===
    private long version;
    private long timestamp;

    // === 构造函数 ===

    public AvatarState() {
        this.facialFeatures = new FacialFeatures();
        this.bodyFeatures = new BodyFeatures();
        this.ageAppearance = new AgeAppearance();
        this.stylePreferences = new StylePreferences();
        this.model3D = new Model3DReference();
        this.displaySettings = new DisplaySettings();
        this.decisionContext = new AvatarDecisionContext();
    }

    // === 面部特征 ===

    @NonNull
    public FacialFeatures getFacialFeatures() {
        return facialFeatures;
    }

    public void setFacialFeatures(@NonNull FacialFeatures facialFeatures) {
        this.facialFeatures = facialFeatures;
    }

    // === 体型特征 ===

    @NonNull
    public BodyFeatures getBodyFeatures() {
        return bodyFeatures;
    }

    public void setBodyFeatures(@NonNull BodyFeatures bodyFeatures) {
        this.bodyFeatures = bodyFeatures;
    }

    // === 年龄外观 ===

    @NonNull
    public AgeAppearance getAgeAppearance() {
        return ageAppearance;
    }

    public void setAgeAppearance(@NonNull AgeAppearance ageAppearance) {
        this.ageAppearance = ageAppearance;
    }

    // === 风格偏好 ===

    @NonNull
    public StylePreferences getStylePreferences() {
        return stylePreferences;
    }

    public void setStylePreferences(@NonNull StylePreferences stylePreferences) {
        this.stylePreferences = stylePreferences;
    }

    // === 3D模型 ===

    @NonNull
    public Model3DReference getModel3D() {
        return model3D;
    }

    public void setModel3D(@NonNull Model3DReference model3D) {
        this.model3D = model3D;
    }

    // === 展示设置 ===

    @NonNull
    public DisplaySettings getDisplaySettings() {
        return displaySettings;
    }

    public void setDisplaySettings(@NonNull DisplaySettings displaySettings) {
        this.displaySettings = displaySettings;
    }

    // === 决策上下文 ===

    @NonNull
    public AvatarDecisionContext getDecisionContext() {
        return decisionContext;
    }

    public void setDecisionContext(@NonNull AvatarDecisionContext decisionContext) {
        this.decisionContext = decisionContext;
    }

    // === 版本信息 ===

    public long getVersion() {
        return version;
    }

    public void setVersion(long version) {
        this.version = version;
    }

    public long getTimestamp() {
        return timestamp;
    }

    public void setTimestamp(long timestamp) {
        this.timestamp = timestamp;
    }

    // === JSON 序列化 ===

    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("facial_features", facialFeatures.toJson());
            json.put("body_features", bodyFeatures.toJson());
            json.put("age_appearance", ageAppearance.toJson());
            json.put("style_preferences", stylePreferences.toJson());
            json.put("model_3d", model3D.toJson());
            json.put("display_settings", displaySettings.toJson());
            json.put("decision_context", decisionContext.toJson());
            json.put("version", version);
            json.put("timestamp", timestamp);
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    @NonNull
    public static AvatarState fromJson(@NonNull JSONObject json) throws JSONException {
        AvatarState state = new AvatarState();

        if (json.has("facial_features")) {
            state.facialFeatures = FacialFeatures.fromJson(json.getJSONObject("facial_features"));
        }
        if (json.has("body_features")) {
            state.bodyFeatures = BodyFeatures.fromJson(json.getJSONObject("body_features"));
        }
        if (json.has("age_appearance")) {
            state.ageAppearance = AgeAppearance.fromJson(json.getJSONObject("age_appearance"));
        }
        if (json.has("style_preferences")) {
            state.stylePreferences = StylePreferences.fromJson(json.getJSONObject("style_preferences"));
        }
        if (json.has("model_3d")) {
            state.model3D = Model3DReference.fromJson(json.getJSONObject("model_3d"));
        }
        if (json.has("display_settings")) {
            state.displaySettings = DisplaySettings.fromJson(json.getJSONObject("display_settings"));
        }
        if (json.has("decision_context")) {
            state.decisionContext = AvatarDecisionContext.fromJson(json.getJSONObject("decision_context"));
        }
        if (json.has("version")) {
            state.version = json.getLong("version");
        }
        if (json.has("timestamp")) {
            state.timestamp = json.getLong("timestamp");
        }

        return state;
    }

    // === 内部模型类 ===

    /**
     * 面部特征
     */
    public static class FacialFeatures {
        public String faceShape;        // oval, round, square, heart
        public String eyeShape;         // almond, round, hooded
        public String eyeColor;         // brown, black, blue, green
        public String skinTone;         // fair, light, medium, dark
        public String hairStyle;        // short, medium, long, curly
        public String hairColor;        // black, brown, blonde
        public double expressiveness;   // 0-1 表现力

        public FacialFeatures() {
            this.faceShape = "oval";
            this.eyeShape = "almond";
            this.eyeColor = "brown";
            this.skinTone = "medium";
            this.hairStyle = "medium";
            this.hairColor = "black";
            this.expressiveness = 0.5;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("face_shape", faceShape);
                json.put("eye_shape", eyeShape);
                json.put("eye_color", eyeColor);
                json.put("skin_tone", skinTone);
                json.put("hair_style", hairStyle);
                json.put("hair_color", hairColor);
                json.put("expressiveness", expressiveness);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static FacialFeatures fromJson(@NonNull JSONObject json) throws JSONException {
            FacialFeatures features = new FacialFeatures();
            if (json.has("face_shape")) features.faceShape = json.getString("face_shape");
            if (json.has("eye_shape")) features.eyeShape = json.getString("eye_shape");
            if (json.has("eye_color")) features.eyeColor = json.getString("eye_color");
            if (json.has("skin_tone")) features.skinTone = json.getString("skin_tone");
            if (json.has("hair_style")) features.hairStyle = json.getString("hair_style");
            if (json.has("hair_color")) features.hairColor = json.getString("hair_color");
            if (json.has("expressiveness")) features.expressiveness = json.getDouble("expressiveness");
            return features;
        }

        /**
         * 是否高表现力
         */
        public boolean isHighExpressiveness() {
            return expressiveness > 0.6;
        }

        /**
         * 是否低表现力
         */
        public boolean isLowExpressiveness() {
            return expressiveness < 0.4;
        }
    }

    /**
     * 体型特征
     */
    public static class BodyFeatures {
        public double height;           // 身高 cm
        public double weight;           // 体重 kg
        public String bodyType;         // slim, average, athletic, curvy
        public String posture;          // confident, modest, casual
        public double postureScore;     // 0-1 姿态质量
        public String movementStyle;    // graceful, energetic, calm
        public String movementSpeed;    // slow, moderate, fast
        public double gestureFrequency; // 0-1 手势频率

        public BodyFeatures() {
            this.height = 170;
            this.weight = 65;
            this.bodyType = "average";
            this.posture = "balanced";
            this.postureScore = 0.5;
            this.movementStyle = "calm";
            this.movementSpeed = "moderate";
            this.gestureFrequency = 0.5;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("height", height);
                json.put("weight", weight);
                json.put("body_type", bodyType);
                json.put("posture", posture);
                json.put("posture_score", postureScore);
                json.put("movement_style", movementStyle);
                json.put("movement_speed", movementSpeed);
                json.put("gesture_frequency", gestureFrequency);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static BodyFeatures fromJson(@NonNull JSONObject json) throws JSONException {
            BodyFeatures features = new BodyFeatures();
            if (json.has("height")) features.height = json.getDouble("height");
            if (json.has("weight")) features.weight = json.getDouble("weight");
            if (json.has("body_type")) features.bodyType = json.getString("body_type");
            if (json.has("posture")) features.posture = json.getString("posture");
            if (json.has("posture_score")) features.postureScore = json.getDouble("posture_score");
            if (json.has("movement_style")) features.movementStyle = json.getString("movement_style");
            if (json.has("movement_speed")) features.movementSpeed = json.getString("movement_speed");
            if (json.has("gesture_frequency")) features.gestureFrequency = json.getDouble("gesture_frequency");
            return features;
        }

        /**
         * 计算 BMI
         */
        public double getBMI() {
            double heightM = height / 100.0;
            if (heightM <= 0) return 0;
            return weight / (heightM * heightM);
        }

        /**
         * 是否自信姿态
         */
        public boolean isConfidentPosture() {
            return "confident".equals(posture) || postureScore > 0.6;
        }

        /**
         * 是否活跃动作
         */
        public boolean isActiveMovement() {
            return "energetic".equals(movementStyle) || gestureFrequency > 0.6;
        }
    }

    /**
     * 年龄外观
     */
    public static class AgeAppearance {
        public int apparentAge;         // 外观年龄
        public String ageRange;         // young, young_adult, adult, middle_aged, senior
        public String agingStage;       // youthful, prime, mature, senior
        public double facialMaturity;   // 0-1 面部成熟度
        public String wrinkleLevel;     // none, minimal, moderate, significant
        public String skinElasticity;   // high, moderate, low

        public AgeAppearance() {
            this.apparentAge = 25;
            this.ageRange = "young_adult";
            this.agingStage = "youthful";
            this.facialMaturity = 0.5;
            this.wrinkleLevel = "minimal";
            this.skinElasticity = "high";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("apparent_age", apparentAge);
                json.put("age_range", ageRange);
                json.put("aging_stage", agingStage);
                json.put("facial_maturity", facialMaturity);
                json.put("wrinkle_level", wrinkleLevel);
                json.put("skin_elasticity", skinElasticity);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static AgeAppearance fromJson(@NonNull JSONObject json) throws JSONException {
            AgeAppearance appearance = new AgeAppearance();
            if (json.has("apparent_age")) appearance.apparentAge = json.getInt("apparent_age");
            if (json.has("age_range")) appearance.ageRange = json.getString("age_range");
            if (json.has("aging_stage")) appearance.agingStage = json.getString("aging_stage");
            if (json.has("facial_maturity")) appearance.facialMaturity = json.getDouble("facial_maturity");
            if (json.has("wrinkle_level")) appearance.wrinkleLevel = json.getString("wrinkle_level");
            if (json.has("skin_elasticity")) appearance.skinElasticity = json.getString("skin_elasticity");
            return appearance;
        }

        /**
         * 是否年轻外观
         */
        public boolean isYouthful() {
            return apparentAge < 30 || "youthful".equals(agingStage);
        }

        /**
         * 是否老年外观
         */
        public boolean isSenior() {
            return apparentAge > 60 || "senior".equals(agingStage) || "elderly".equals(agingStage);
        }
    }

    /**
     * 风格偏好
     */
    public static class StylePreferences {
        public String clothingStyle;    // casual, business, sporty, elegant
        public String clothingQuality;  // budget, mid_range, premium, luxury
        public List<String> clothingColors;
        public String accessoryStyle;   // minimal, moderate, bold
        public String groomingLevel;    // natural, polished, elaborate
        public String overallVibe;      // professional, casual, artistic
        public String aestheticTheme;   // classic, modern, vintage
        public String culturalStyle;    // traditional, modern, fusion
        public double brandAwareness;   // 0-1 品牌意识

        public StylePreferences() {
            this.clothingStyle = "casual";
            this.clothingQuality = "mid_range";
            this.clothingColors = new ArrayList<>();
            this.accessoryStyle = "minimal";
            this.groomingLevel = "natural";
            this.overallVibe = "balanced";
            this.aestheticTheme = "modern";
            this.culturalStyle = "modern";
            this.brandAwareness = 0.3;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("clothing_style", clothingStyle);
                json.put("clothing_quality", clothingQuality);
                json.put("accessory_style", accessoryStyle);
                json.put("grooming_level", groomingLevel);
                json.put("overall_vibe", overallVibe);
                json.put("aesthetic_theme", aestheticTheme);
                json.put("cultural_style", culturalStyle);
                json.put("brand_awareness", brandAwareness);

                JSONArray colorsArray = new JSONArray();
                for (String color : clothingColors) {
                    colorsArray.put(color);
                }
                json.put("clothing_colors", colorsArray);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static StylePreferences fromJson(@NonNull JSONObject json) throws JSONException {
            StylePreferences preferences = new StylePreferences();
            if (json.has("clothing_style")) preferences.clothingStyle = json.getString("clothing_style");
            if (json.has("clothing_quality")) preferences.clothingQuality = json.getString("clothing_quality");
            if (json.has("accessory_style")) preferences.accessoryStyle = json.getString("accessory_style");
            if (json.has("grooming_level")) preferences.groomingLevel = json.getString("grooming_level");
            if (json.has("overall_vibe")) preferences.overallVibe = json.getString("overall_vibe");
            if (json.has("aesthetic_theme")) preferences.aestheticTheme = json.getString("aesthetic_theme");
            if (json.has("cultural_style")) preferences.culturalStyle = json.getString("cultural_style");
            if (json.has("brand_awareness")) preferences.brandAwareness = json.getDouble("brand_awareness");

            if (json.has("clothing_colors")) {
                JSONArray colorsArray = json.getJSONArray("clothing_colors");
                for (int i = 0; i < colorsArray.length(); i++) {
                    preferences.clothingColors.add(colorsArray.getString(i));
                }
            }

            return preferences;
        }

        /**
         * 是否正式风格
         */
        public boolean isFormalStyle() {
            return "business".equals(clothingStyle) || "formal".equals(clothingStyle);
        }

        /**
         * 是否高品牌意识
         */
        public boolean isBrandConscious() {
            return brandAwareness > 0.6;
        }
    }

    /**
     * 3D模型引用
     */
    public static class Model3DReference {
        public String modelId;          // 模型ID
        public String modelType;        // custom, preset, generated
        public String sourceFormat;     // glb, gltf, fbx
        public boolean animationEnabled;
        public List<String> animationSet;
        public String renderQuality;    // low, medium, high
        public String optimizedFor;     // mobile, desktop, vr

        public Model3DReference() {
            this.modelType = "preset";
            this.sourceFormat = "glb";
            this.animationEnabled = true;
            this.animationSet = new ArrayList<>();
            this.renderQuality = "medium";
            this.optimizedFor = "mobile";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("model_id", modelId);
                json.put("model_type", modelType);
                json.put("source_format", sourceFormat);
                json.put("animation_enabled", animationEnabled);
                json.put("render_quality", renderQuality);
                json.put("optimized_for", optimizedFor);

                JSONArray animArray = new JSONArray();
                for (String anim : animationSet) {
                    animArray.put(anim);
                }
                json.put("animation_set", animArray);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static Model3DReference fromJson(@NonNull JSONObject json) throws JSONException {
            Model3DReference ref = new Model3DReference();
            if (json.has("model_id")) ref.modelId = json.getString("model_id");
            if (json.has("model_type")) ref.modelType = json.getString("model_type");
            if (json.has("source_format")) ref.sourceFormat = json.getString("source_format");
            if (json.has("animation_enabled")) ref.animationEnabled = json.getBoolean("animation_enabled");
            if (json.has("render_quality")) ref.renderQuality = json.getString("render_quality");
            if (json.has("optimized_for")) ref.optimizedFor = json.getString("optimized_for");

            if (json.has("animation_set")) {
                JSONArray animArray = json.getJSONArray("animation_set");
                for (int i = 0; i < animArray.length(); i++) {
                    ref.animationSet.add(animArray.getString(i));
                }
            }

            return ref;
        }

        /**
         * 是否支持动画
         */
        public boolean supportsAnimation() {
            return animationEnabled && !animationSet.isEmpty();
        }
    }

    /**
     * 展示设置
     */
    public static class DisplaySettings {
        public String renderMode;       // 2d, 3d, vr, ar
        public String cameraPosition;   // front, side, portrait
        public String cameraDistance;   // close, medium, far
        public String background;       // transparent, neutral, scene
        public String animationState;   // idle, speaking, walking
        public String expression;       // neutral, happy, serious

        public DisplaySettings() {
            this.renderMode = "3d";
            this.cameraPosition = "portrait";
            this.cameraDistance = "medium";
            this.background = "neutral";
            this.animationState = "idle";
            this.expression = "neutral";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("render_mode", renderMode);
                json.put("camera_position", cameraPosition);
                json.put("camera_distance", cameraDistance);
                json.put("background", background);
                json.put("animation_state", animationState);
                json.put("expression", expression);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static DisplaySettings fromJson(@NonNull JSONObject json) throws JSONException {
            DisplaySettings settings = new DisplaySettings();
            if (json.has("render_mode")) settings.renderMode = json.getString("render_mode");
            if (json.has("camera_position")) settings.cameraPosition = json.getString("camera_position");
            if (json.has("camera_distance")) settings.cameraDistance = json.getString("camera_distance");
            if (json.has("background")) settings.background = json.getString("background");
            if (json.has("animation_state")) settings.animationState = json.getString("animation_state");
            if (json.has("expression")) settings.expression = json.getString("expression");
            return settings;
        }

        /**
         * 是否 3D 模式
         */
        public boolean is3DMode() {
            return "3d".equals(renderMode) || "vr".equals(renderMode);
        }

        /**
         * 是否 VR 模式
         */
        public boolean isVRMode() {
            return "vr".equals(renderMode);
        }
    }

    /**
     * Avatar 决策上下文
     */
    public static class AvatarDecisionContext {
        public String recommendedStyle;       // 推荐风格
        public String recommendedPosture;     // 推荐姿态
        public String recommendedExpression;  // 推荐表情
        public SceneAdaptation sceneAdaptation;
        public SocialAdaptation socialAdaptation;
        public CulturalAdaptation culturalAdaptation;

        public AvatarDecisionContext() {
            this.recommendedStyle = "casual";
            this.recommendedPosture = "balanced";
            this.recommendedExpression = "neutral";
            this.sceneAdaptation = new SceneAdaptation();
            this.socialAdaptation = new SocialAdaptation();
            this.culturalAdaptation = new CulturalAdaptation();
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("recommended_style", recommendedStyle);
                json.put("recommended_posture", recommendedPosture);
                json.put("recommended_expression", recommendedExpression);
                json.put("scene_adaptation", sceneAdaptation.toJson());
                json.put("social_adaptation", socialAdaptation.toJson());
                json.put("cultural_adaptation", culturalAdaptation.toJson());
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static AvatarDecisionContext fromJson(@NonNull JSONObject json) throws JSONException {
            AvatarDecisionContext ctx = new AvatarDecisionContext();
            if (json.has("recommended_style")) ctx.recommendedStyle = json.getString("recommended_style");
            if (json.has("recommended_posture")) ctx.recommendedPosture = json.getString("recommended_posture");
            if (json.has("recommended_expression")) ctx.recommendedExpression = json.getString("recommended_expression");
            if (json.has("scene_adaptation")) ctx.sceneAdaptation = SceneAdaptation.fromJson(json.getJSONObject("scene_adaptation"));
            if (json.has("social_adaptation")) ctx.socialAdaptation = SocialAdaptation.fromJson(json.getJSONObject("social_adaptation"));
            if (json.has("cultural_adaptation")) ctx.culturalAdaptation = CulturalAdaptation.fromJson(json.getJSONObject("cultural_adaptation"));
            return ctx;
        }
    }

    /**
     * 场景适应
     */
    public static class SceneAdaptation {
        public String currentScene;         // meeting, casual, formal, sport
        public String styleAdjustment;      // formal_up, casual_down, neutral
        public String postureAdjustment;    // confident, relaxed, formal
        public String expressionRange;      // professional, warm, neutral
        public String animationSet;         // idle_meeting, idle_casual

        public SceneAdaptation() {
            this.currentScene = "casual";
            this.styleAdjustment = "neutral";
            this.postureAdjustment = "balanced";
            this.expressionRange = "neutral";
            this.animationSet = "idle";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("current_scene", currentScene);
                json.put("style_adjustment", styleAdjustment);
                json.put("posture_adjustment", postureAdjustment);
                json.put("expression_range", expressionRange);
                json.put("animation_set", animationSet);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static SceneAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            SceneAdaptation sa = new SceneAdaptation();
            if (json.has("current_scene")) sa.currentScene = json.getString("current_scene");
            if (json.has("style_adjustment")) sa.styleAdjustment = json.getString("style_adjustment");
            if (json.has("posture_adjustment")) sa.postureAdjustment = json.getString("posture_adjustment");
            if (json.has("expression_range")) sa.expressionRange = json.getString("expression_range");
            if (json.has("animation_set")) sa.animationSet = json.getString("animation_set");
            return sa;
        }

        /**
         * 是否正式场景
         */
        public boolean isFormalScene() {
            return "meeting".equals(currentScene) || "formal".equals(currentScene);
        }
    }

    /**
     * 社交适应
     */
    public static class SocialAdaptation {
        public String socialContext;        // formal, casual, intimate
        public String distanceLevel;        // close, moderate, far
        public double eyeContactLevel;      // 0-1 眼神接触程度
        public double gestureLevel;         // 0-1 手势程度
        public String touchPermission;      // none, handshake, hug

        public SocialAdaptation() {
            this.socialContext = "casual";
            this.distanceLevel = "moderate";
            this.eyeContactLevel = 0.5;
            this.gestureLevel = 0.5;
            this.touchPermission = "handshake";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("social_context", socialContext);
                json.put("distance_level", distanceLevel);
                json.put("eye_contact_level", eyeContactLevel);
                json.put("gesture_level", gestureLevel);
                json.put("touch_permission", touchPermission);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static SocialAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            SocialAdaptation sa = new SocialAdaptation();
            if (json.has("social_context")) sa.socialContext = json.getString("social_context");
            if (json.has("distance_level")) sa.distanceLevel = json.getString("distance_level");
            if (json.has("eye_contact_level")) sa.eyeContactLevel = json.getDouble("eye_contact_level");
            if (json.has("gesture_level")) sa.gestureLevel = json.getDouble("gesture_level");
            if (json.has("touch_permission")) sa.touchPermission = json.getString("touch_permission");
            return sa;
        }

        /**
         * 是否亲密社交
         */
        public boolean isIntimateContext() {
            return "intimate".equals(socialContext) || "close".equals(distanceLevel);
        }
    }

    /**
     * 文化适应
     */
    public static class CulturalAdaptation {
        public String culturalContext;      // local, international, multicultural
        public double formalityLevel;       // 0-1 正式程度
        public double modestyLevel;         // 0-1 谦逊程度
        public String greetingStyle;        // handshake, bow, wave
        public String communicationStyle;   // direct, indirect

        public CulturalAdaptation() {
            this.culturalContext = "international";
            this.formalityLevel = 0.5;
            this.modestyLevel = 0.5;
            this.greetingStyle = "handshake";
            this.communicationStyle = "context_dependent";
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("cultural_context", culturalContext);
                json.put("formality_level", formalityLevel);
                json.put("modesty_level", modestyLevel);
                json.put("greeting_style", greetingStyle);
                json.put("communication_style", communicationStyle);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        @NonNull
        public static CulturalAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            CulturalAdaptation ca = new CulturalAdaptation();
            if (json.has("cultural_context")) ca.culturalContext = json.getString("cultural_context");
            if (json.has("formality_level")) ca.formalityLevel = json.getDouble("formality_level");
            if (json.has("modesty_level")) ca.modestyLevel = json.getDouble("modesty_level");
            if (json.has("greeting_style")) ca.greetingStyle = json.getString("greeting_style");
            if (json.has("communication_style")) ca.communicationStyle = json.getString("communication_style");
            return ca;
        }

        /**
         * 是否高正式度
         */
        public boolean isHighFormality() {
            return formalityLevel > 0.7;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "AvatarState{" +
                "age=" + ageAppearance.apparentAge +
                ", bodyType='" + bodyFeatures.bodyType + '\'' +
                ", style='" + stylePreferences.clothingStyle + '\'' +
                ", posture='" + bodyFeatures.posture + '\'' +
                '}';
    }
}