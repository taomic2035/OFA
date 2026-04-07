package com.ofa.agent.expression;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * ExpressionGesture状态模型 (v5.3.0)
 *
 * 端侧接收 Center 推送的表情动作状态，用于动画渲染和表达。
 * 深层表情动作管理在 Center 端 ExpressionGestureEngine 完成。
 */
public class ExpressionGestureState {

    // === Facial Expression Settings ===
    private FacialExpressionSettings facialExpressionSettings;

    // === Body Gesture Settings ===
    private BodyGestureSettings bodyGestureSettings;

    // === Emotion Expression Mapping ===
    private EmotionExpressionMapping emotionExpressionMapping;

    // === Social Gesture Settings ===
    private SocialGestureSettings socialGestureSettings;

    // === Animation Settings ===
    private AnimationSettings animationSettings;

    // === Expression Gesture Context ===
    private ExpressionGestureContext context;

    public ExpressionGestureState() {
        this.facialExpressionSettings = new FacialExpressionSettings();
        this.bodyGestureSettings = new BodyGestureSettings();
        this.emotionExpressionMapping = new EmotionExpressionMapping();
        this.socialGestureSettings = new SocialGestureSettings();
        this.animationSettings = new AnimationSettings();
        this.context = new ExpressionGestureContext();
    }

    // === Getters ===

    @NonNull
    public FacialExpressionSettings getFacialExpressionSettings() {
        return facialExpressionSettings;
    }

    @NonNull
    public BodyGestureSettings getBodyGestureSettings() {
        return bodyGestureSettings;
    }

    @NonNull
    public EmotionExpressionMapping getEmotionExpressionMapping() {
        return emotionExpressionMapping;
    }

    @NonNull
    public SocialGestureSettings getSocialGestureSettings() {
        return socialGestureSettings;
    }

    @NonNull
    public AnimationSettings getAnimationSettings() {
        return animationSettings;
    }

    @NonNull
    public ExpressionGestureContext getContext() {
        return context;
    }

    // === Facial Expression Settings ===

    public static class FacialExpressionSettings {
        public String defaultExpression = "neutral";
        public double expressionRange = 0.5;
        public double expressionIntensity = 0.5;
        public double expressionFrequency = 0.3;
        public String expressionDuration = "medium";

        public boolean eyeExpressionEnabled = true;
        public double eyeContactTendency = 0.5;
        public double blinkRate = 15.0;
        public double eyebrowExpressiveness = 0.5;

        public boolean mouthExpressionEnabled = true;
        public double smileTendency = 0.5;
        public String smileType = "moderate";
        public double lipMovementExpressiveness = 0.5;

        public boolean microExpressionEnabled = true;
        public double microExpressionSensitivity = 0.5;
        public int microExpressionDuration = 200;

        public double symmetryLevel = 0.8;
        public double expressionMasking = 0.0;
        public double pokerFaceAbility = 0.0;

        public boolean isHighExpressiveness() {
            return expressionIntensity > 0.6;
        }

        public boolean isSubtleExpression() {
            return expressionIntensity < 0.4;
        }

        public boolean isNaturalBlink() {
            return blinkRate >= 12 && blinkRate <= 20;
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("default_expression", defaultExpression);
                json.put("expression_range", expressionRange);
                json.put("expression_intensity", expressionIntensity);
                json.put("expression_frequency", expressionFrequency);
                json.put("expression_duration", expressionDuration);
                json.put("eye_expression_enabled", eyeExpressionEnabled);
                json.put("eye_contact_tendency", eyeContactTendency);
                json.put("blink_rate", blinkRate);
                json.put("eyebrow_expressiveness", eyebrowExpressiveness);
                json.put("mouth_expression_enabled", mouthExpressionEnabled);
                json.put("smile_tendency", smileTendency);
                json.put("smile_type", smileType);
                json.put("lip_movement_expressiveness", lipMovementExpressiveness);
                json.put("micro_expression_enabled", microExpressionEnabled);
                json.put("micro_expression_sensitivity", microExpressionSensitivity);
                json.put("micro_expression_duration", microExpressionDuration);
                json.put("symmetry_level", symmetryLevel);
                json.put("expression_masking", expressionMasking);
                json.put("poker_face_ability", pokerFaceAbility);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static FacialExpressionSettings fromJson(@NonNull JSONObject json) throws JSONException {
            FacialExpressionSettings settings = new FacialExpressionSettings();
            settings.defaultExpression = json.optString("default_expression", "neutral");
            settings.expressionRange = json.optDouble("expression_range", 0.5);
            settings.expressionIntensity = json.optDouble("expression_intensity", 0.5);
            settings.expressionFrequency = json.optDouble("expression_frequency", 0.3);
            settings.expressionDuration = json.optString("expression_duration", "medium");
            settings.eyeExpressionEnabled = json.optBoolean("eye_expression_enabled", true);
            settings.eyeContactTendency = json.optDouble("eye_contact_tendency", 0.5);
            settings.blinkRate = json.optDouble("blink_rate", 15.0);
            settings.eyebrowExpressiveness = json.optDouble("eyebrow_expressiveness", 0.5);
            settings.mouthExpressionEnabled = json.optBoolean("mouth_expression_enabled", true);
            settings.smileTendency = json.optDouble("smile_tendency", 0.5);
            settings.smileType = json.optString("smile_type", "moderate");
            settings.lipMovementExpressiveness = json.optDouble("lip_movement_expressiveness", 0.5);
            settings.microExpressionEnabled = json.optBoolean("micro_expression_enabled", true);
            settings.microExpressionSensitivity = json.optDouble("micro_expression_sensitivity", 0.5);
            settings.microExpressionDuration = json.optInt("micro_expression_duration", 200);
            settings.symmetryLevel = json.optDouble("symmetry_level", 0.8);
            settings.expressionMasking = json.optDouble("expression_masking", 0.0);
            settings.pokerFaceAbility = json.optDouble("poker_face_ability", 0.0);
            return settings;
        }
    }

    // === Body Gesture Settings ===

    public static class BodyGestureSettings {
        public String defaultPosture = "neutral";
        public double gestureRange = 0.5;
        public double gestureIntensity = 0.5;
        public double gestureFrequency = 0.3;
        public String gestureSpeed = "moderate";

        public boolean handGestureEnabled = true;
        public String handGestureStyle = "moderate";
        public String handPosition = "natural";
        public List<String> handGestureVocabulary = new ArrayList<>();

        public boolean headMovementEnabled = true;
        public double nodFrequency = 0.3;
        public double headTiltTendency = 0.3;
        public double headShakeFrequency = 0.1;

        public boolean bodyLeanEnabled = true;
        public String bodyLeanDirection = "neutral";
        public double bodyLeanTendency = 0.3;

        public double shrugTendency = 0.1;
        public String shoulderTension = "relaxed";

        public double fidgetLevel = 0.0;
        public String fidgetType = "none";

        public boolean mirroringEnabled = false;
        public int mirroringDelay = 500;
        public double mirroringIntensity = 0.5;

        public boolean isExpressiveGestures() {
            return gestureIntensity > 0.6;
        }

        public boolean isMinimalGestures() {
            return gestureIntensity < 0.4;
        }

        public boolean isNaturalPosture() {
            return "neutral".equals(defaultPosture) || "relaxed".equals(defaultPosture);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("default_posture", defaultPosture);
                json.put("gesture_range", gestureRange);
                json.put("gesture_intensity", gestureIntensity);
                json.put("gesture_frequency", gestureFrequency);
                json.put("gesture_speed", gestureSpeed);
                json.put("hand_gesture_enabled", handGestureEnabled);
                json.put("hand_gesture_style", handGestureStyle);
                json.put("hand_position", handPosition);
                json.put("hand_gesture_vocabulary", new JSONArray(handGestureVocabulary));
                json.put("head_movement_enabled", headMovementEnabled);
                json.put("nod_frequency", nodFrequency);
                json.put("head_tilt_tendency", headTiltTendency);
                json.put("head_shake_frequency", headShakeFrequency);
                json.put("body_lean_enabled", bodyLeanEnabled);
                json.put("body_lean_direction", bodyLeanDirection);
                json.put("body_lean_tendency", bodyLeanTendency);
                json.put("shrug_tendency", shrugTendency);
                json.put("shoulder_tension", shoulderTension);
                json.put("fidget_level", fidgetLevel);
                json.put("fidget_type", fidgetType);
                json.put("mirroring_enabled", mirroringEnabled);
                json.put("mirroring_delay", mirroringDelay);
                json.put("mirroring_intensity", mirroringIntensity);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static BodyGestureSettings fromJson(@NonNull JSONObject json) throws JSONException {
            BodyGestureSettings settings = new BodyGestureSettings();
            settings.defaultPosture = json.optString("default_posture", "neutral");
            settings.gestureRange = json.optDouble("gesture_range", 0.5);
            settings.gestureIntensity = json.optDouble("gesture_intensity", 0.5);
            settings.gestureFrequency = json.optDouble("gesture_frequency", 0.3);
            settings.gestureSpeed = json.optString("gesture_speed", "moderate");
            settings.handGestureEnabled = json.optBoolean("hand_gesture_enabled", true);
            settings.handGestureStyle = json.optString("hand_gesture_style", "moderate");
            settings.handPosition = json.optString("hand_position", "natural");
            JSONArray vocabArray = json.optJSONArray("hand_gesture_vocabulary");
            if (vocabArray != null) {
                for (int i = 0; i < vocabArray.length(); i++) {
                    settings.handGestureVocabulary.add(vocabArray.getString(i));
                }
            }
            settings.headMovementEnabled = json.optBoolean("head_movement_enabled", true);
            settings.nodFrequency = json.optDouble("nod_frequency", 0.3);
            settings.headTiltTendency = json.optDouble("head_tilt_tendency", 0.3);
            settings.headShakeFrequency = json.optDouble("head_shake_frequency", 0.1);
            settings.bodyLeanEnabled = json.optBoolean("body_lean_enabled", true);
            settings.bodyLeanDirection = json.optString("body_lean_direction", "neutral");
            settings.bodyLeanTendency = json.optDouble("body_lean_tendency", 0.3);
            settings.shrugTendency = json.optDouble("shrug_tendency", 0.1);
            settings.shoulderTension = json.optString("shoulder_tension", "relaxed");
            settings.fidgetLevel = json.optDouble("fidget_level", 0.0);
            settings.fidgetType = json.optString("fidget_type", "none");
            settings.mirroringEnabled = json.optBoolean("mirroring_enabled", false);
            settings.mirroringDelay = json.optInt("mirroring_delay", 500);
            settings.mirroringIntensity = json.optDouble("mirroring_intensity", 0.5);
            return settings;
        }
    }

    // === Emotion Expression Mapping ===

    public static class EmotionExpressionMapping {
        public Map<String, ExpressionMapping> emotionMappings = new HashMap<>();
        public String transitionSpeed = "smooth";
        public int transitionDuration = 300;
        public boolean blendEnabled = true;
        public int blendDuration = 200;

        @Nullable
        public ExpressionMapping getMappingForEmotion(@NonNull String emotion) {
            return emotionMappings.get(emotion);
        }

        public boolean hasMapping(@NonNull String emotion) {
            return emotionMappings.containsKey(emotion);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                JSONObject mappingsJson = new JSONObject();
                for (Map.Entry<String, ExpressionMapping> entry : emotionMappings.entrySet()) {
                    mappingsJson.put(entry.getKey(), entry.getValue().toJson());
                }
                json.put("emotion_mappings", mappingsJson);
                json.put("transition_speed", transitionSpeed);
                json.put("transition_duration", transitionDuration);
                json.put("blend_enabled", blendEnabled);
                json.put("blend_duration", blendDuration);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static EmotionExpressionMapping fromJson(@NonNull JSONObject json) throws JSONException {
            EmotionExpressionMapping mapping = new EmotionExpressionMapping();
            mapping.transitionSpeed = json.optString("transition_speed", "smooth");
            mapping.transitionDuration = json.optInt("transition_duration", 300);
            mapping.blendEnabled = json.optBoolean("blend_enabled", true);
            mapping.blendDuration = json.optInt("blend_duration", 200);
            JSONObject mappingsJson = json.optJSONObject("emotion_mappings");
            if (mappingsJson != null) {
                JSONArray names = mappingsJson.names();
                if (names != null) {
                    for (int i = 0; i < names.length(); i++) {
                        String key = names.getString(i);
                        JSONObject mappingJson = mappingsJson.getJSONObject(key);
                        mapping.emotionMappings.put(key, ExpressionMapping.fromJson(mappingJson));
                    }
                }
            }
            return mapping;
        }
    }

    // === Expression Mapping ===

    public static class ExpressionMapping {
        public String emotionName;

        // Facial expression
        public String eyebrowPosition = "neutral";
        public String eyeShape = "normal";
        public String mouthShape = "neutral";
        public String expressionType = "neutral";
        public double intensity = 0.5;

        // Body gesture
        public String posture = "neutral";
        public String handGesture = "none";
        public String bodyMovement = "none";

        // Micro-expression
        public String microExpression = "none";

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("emotion_name", emotionName);
                json.put("eyebrow_position", eyebrowPosition);
                json.put("eye_shape", eyeShape);
                json.put("mouth_shape", mouthShape);
                json.put("expression_type", expressionType);
                json.put("intensity", intensity);
                json.put("posture", posture);
                json.put("hand_gesture", handGesture);
                json.put("body_movement", bodyMovement);
                json.put("micro_expression", microExpression);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ExpressionMapping fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionMapping mapping = new ExpressionMapping();
            mapping.emotionName = json.optString("emotion_name", "");
            mapping.eyebrowPosition = json.optString("eyebrow_position", "neutral");
            mapping.eyeShape = json.optString("eye_shape", "normal");
            mapping.mouthShape = json.optString("mouth_shape", "neutral");
            mapping.expressionType = json.optString("expression_type", "neutral");
            mapping.intensity = json.optDouble("intensity", 0.5);
            mapping.posture = json.optString("posture", "neutral");
            mapping.handGesture = json.optString("hand_gesture", "none");
            mapping.bodyMovement = json.optString("body_movement", "none");
            mapping.microExpression = json.optString("micro_expression", "none");
            return mapping;
        }
    }

    // === Social Gesture Settings ===

    public static class SocialGestureSettings {
        public String greetingGesture = "nod";
        public double greetingIntensity = 0.5;

        public String partingGesture = "nod";
        public double partingIntensity = 0.5;

        public boolean listeningGestureEnabled = true;
        public double nodWhileListening = 0.5;
        public double eyeContactWhileListening = 0.6;

        public boolean speakingGestureEnabled = true;
        public double gestureWhileSpeaking = 0.5;
        public String pauseGesture = "none";

        public String agreementGesture = "nod";
        public String disagreementGesture = "head_shake";
        public String uncertaintyGesture = "shrug";

        public double touchComfortLevel = 0.3;
        public List<String> touchTypes = new ArrayList<>();

        public String preferredDistance = "medium";
        public double distanceAdjustment = 0.5;

        public boolean mirrorCloseFriends = true;
        public boolean mirrorProfessional = false;
        public boolean mirrorStrangers = false;

        public boolean isComfortableWithTouch() {
            return touchComfortLevel > 0.5;
        }

        public boolean prefersCloseDistance() {
            return "close".equals(preferredDistance);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("greeting_gesture", greetingGesture);
                json.put("greeting_intensity", greetingIntensity);
                json.put("parting_gesture", partingGesture);
                json.put("parting_intensity", partingIntensity);
                json.put("listening_gesture_enabled", listeningGestureEnabled);
                json.put("nod_while_listening", nodWhileListening);
                json.put("eye_contact_while_listening", eyeContactWhileListening);
                json.put("speaking_gesture_enabled", speakingGestureEnabled);
                json.put("gesture_while_speaking", gestureWhileSpeaking);
                json.put("pause_gesture", pauseGesture);
                json.put("agreement_gesture", agreementGesture);
                json.put("disagreement_gesture", disagreementGesture);
                json.put("uncertainty_gesture", uncertaintyGesture);
                json.put("touch_comfort_level", touchComfortLevel);
                json.put("touch_types", new JSONArray(touchTypes));
                json.put("preferred_distance", preferredDistance);
                json.put("distance_adjustment", distanceAdjustment);
                json.put("mirror_close_friends", mirrorCloseFriends);
                json.put("mirror_professional", mirrorProfessional);
                json.put("mirror_strangers", mirrorStrangers);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static SocialGestureSettings fromJson(@NonNull JSONObject json) throws JSONException {
            SocialGestureSettings settings = new SocialGestureSettings();
            settings.greetingGesture = json.optString("greeting_gesture", "nod");
            settings.greetingIntensity = json.optDouble("greeting_intensity", 0.5);
            settings.partingGesture = json.optString("parting_gesture", "nod");
            settings.partingIntensity = json.optDouble("parting_intensity", 0.5);
            settings.listeningGestureEnabled = json.optBoolean("listening_gesture_enabled", true);
            settings.nodWhileListening = json.optDouble("nod_while_listening", 0.5);
            settings.eyeContactWhileListening = json.optDouble("eye_contact_while_listening", 0.6);
            settings.speakingGestureEnabled = json.optBoolean("speaking_gesture_enabled", true);
            settings.gestureWhileSpeaking = json.optDouble("gesture_while_speaking", 0.5);
            settings.pauseGesture = json.optString("pause_gesture", "none");
            settings.agreementGesture = json.optString("agreement_gesture", "nod");
            settings.disagreementGesture = json.optString("disagreement_gesture", "head_shake");
            settings.uncertaintyGesture = json.optString("uncertainty_gesture", "shrug");
            settings.touchComfortLevel = json.optDouble("touch_comfort_level", 0.3);
            JSONArray touchArray = json.optJSONArray("touch_types");
            if (touchArray != null) {
                for (int i = 0; i < touchArray.length(); i++) {
                    settings.touchTypes.add(touchArray.getString(i));
                }
            }
            settings.preferredDistance = json.optString("preferred_distance", "medium");
            settings.distanceAdjustment = json.optDouble("distance_adjustment", 0.5);
            settings.mirrorCloseFriends = json.optBoolean("mirror_close_friends", true);
            settings.mirrorProfessional = json.optBoolean("mirror_professional", false);
            settings.mirrorStrangers = json.optBoolean("mirror_strangers", false);
            return settings;
        }
    }

    // === Animation Settings ===

    public static class AnimationSettings {
        public String animationStyle = "realistic";

        public boolean idleAnimationEnabled = true;
        public List<String> idleAnimationSet = new ArrayList<>();
        public double idleVariationFrequency = 0.3;

        public boolean transitionAnimationsEnabled = true;
        public String transitionSpeed = "normal";

        public String expressionAnimationQuality = "medium";
        public int expressionAnimationFPS = 30;

        public String gestureAnimationQuality = "medium";
        public int gestureAnimationFPS = 30;

        public boolean lipSyncEnabled = true;
        public String lipSyncQuality = "standard";
        public int lipSyncDelay = 0;

        public boolean blinkAnimationEnabled = true;
        public boolean blinkAnimationNatural = true;

        public boolean breathingAnimationEnabled = true;
        public double breathingRate = 16.0;
        public double breathingDepth = 0.5;

        public boolean eyeMovementEnabled = true;
        public double saccadeFrequency = 3.0;
        public double saccadeRange = 0.3;

        public boolean isHighQualityAnimation() {
            return "high".equals(expressionAnimationQuality) && "high".equals(gestureAnimationQuality);
        }

        public boolean isLipSyncHighQuality() {
            return "high".equals(lipSyncQuality);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("animation_style", animationStyle);
                json.put("idle_animation_enabled", idleAnimationEnabled);
                json.put("idle_animation_set", new JSONArray(idleAnimationSet));
                json.put("idle_variation_frequency", idleVariationFrequency);
                json.put("transition_animations_enabled", transitionAnimationsEnabled);
                json.put("transition_speed", transitionSpeed);
                json.put("expression_animation_quality", expressionAnimationQuality);
                json.put("expression_animation_fps", expressionAnimationFPS);
                json.put("gesture_animation_quality", gestureAnimationQuality);
                json.put("gesture_animation_fps", gestureAnimationFPS);
                json.put("lip_sync_enabled", lipSyncEnabled);
                json.put("lip_sync_quality", lipSyncQuality);
                json.put("lip_sync_delay", lipSyncDelay);
                json.put("blink_animation_enabled", blinkAnimationEnabled);
                json.put("blink_animation_natural", blinkAnimationNatural);
                json.put("breathing_animation_enabled", breathingAnimationEnabled);
                json.put("breathing_rate", breathingRate);
                json.put("breathing_depth", breathingDepth);
                json.put("eye_movement_enabled", eyeMovementEnabled);
                json.put("saccade_frequency", saccadeFrequency);
                json.put("saccade_range", saccadeRange);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static AnimationSettings fromJson(@NonNull JSONObject json) throws JSONException {
            AnimationSettings settings = new AnimationSettings();
            settings.animationStyle = json.optString("animation_style", "realistic");
            settings.idleAnimationEnabled = json.optBoolean("idle_animation_enabled", true);
            JSONArray idleArray = json.optJSONArray("idle_animation_set");
            if (idleArray != null) {
                for (int i = 0; i < idleArray.length(); i++) {
                    settings.idleAnimationSet.add(idleArray.getString(i));
                }
            }
            settings.idleVariationFrequency = json.optDouble("idle_variation_frequency", 0.3);
            settings.transitionAnimationsEnabled = json.optBoolean("transition_animations_enabled", true);
            settings.transitionSpeed = json.optString("transition_speed", "normal");
            settings.expressionAnimationQuality = json.optString("expression_animation_quality", "medium");
            settings.expressionAnimationFPS = json.optInt("expression_animation_fps", 30);
            settings.gestureAnimationQuality = json.optString("gesture_animation_quality", "medium");
            settings.gestureAnimationFPS = json.optInt("gesture_animation_fps", 30);
            settings.lipSyncEnabled = json.optBoolean("lip_sync_enabled", true);
            settings.lipSyncQuality = json.optString("lip_sync_quality", "standard");
            settings.lipSyncDelay = json.optInt("lip_sync_delay", 0);
            settings.blinkAnimationEnabled = json.optBoolean("blink_animation_enabled", true);
            settings.blinkAnimationNatural = json.optBoolean("blink_animation_natural", true);
            settings.breathingAnimationEnabled = json.optBoolean("breathing_animation_enabled", true);
            settings.breathingRate = json.optDouble("breathing_rate", 16.0);
            settings.breathingDepth = json.optDouble("breathing_depth", 0.5);
            settings.eyeMovementEnabled = json.optBoolean("eye_movement_enabled", true);
            settings.saccadeFrequency = json.optDouble("saccade_frequency", 3.0);
            settings.saccadeRange = json.optDouble("saccade_range", 0.3);
            return settings;
        }
    }

    // === Expression Gesture Context ===

    public static class ExpressionGestureContext {
        public String identityId;

        public ExpressionState currentExpression = new ExpressionState();
        public GestureState currentGesture = new GestureState();

        public ExpressionState recommendedExpression = new ExpressionState();
        public GestureState recommendedGesture = new GestureState();

        public ExpressionSceneAdaptation sceneAdaptation = new ExpressionSceneAdaptation();
        public ExpressionEmotionAdaptation emotionAdaptation = new ExpressionEmotionAdaptation();
        public ExpressionSocialAdaptation socialAdaptation = new ExpressionSocialAdaptation();

        public AnimationState animationState = new AnimationState();

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("identity_id", identityId);
                json.put("current_expression", currentExpression.toJson());
                json.put("current_gesture", currentGesture.toJson());
                json.put("recommended_expression", recommendedExpression.toJson());
                json.put("recommended_gesture", recommendedGesture.toJson());
                json.put("scene_adaptation", sceneAdaptation.toJson());
                json.put("emotion_adaptation", emotionAdaptation.toJson());
                json.put("social_adaptation", socialAdaptation.toJson());
                json.put("animation_state", animationState.toJson());
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ExpressionGestureContext fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionGestureContext context = new ExpressionGestureContext();
            context.identityId = json.optString("identity_id", "");
            JSONObject currentExprJson = json.optJSONObject("current_expression");
            if (currentExprJson != null) {
                context.currentExpression = ExpressionState.fromJson(currentExprJson);
            }
            JSONObject currentGestureJson = json.optJSONObject("current_gesture");
            if (currentGestureJson != null) {
                context.currentGesture = GestureState.fromJson(currentGestureJson);
            }
            JSONObject recExprJson = json.optJSONObject("recommended_expression");
            if (recExprJson != null) {
                context.recommendedExpression = ExpressionState.fromJson(recExprJson);
            }
            JSONObject recGestureJson = json.optJSONObject("recommended_gesture");
            if (recGestureJson != null) {
                context.recommendedGesture = GestureState.fromJson(recGestureJson);
            }
            JSONObject sceneJson = json.optJSONObject("scene_adaptation");
            if (sceneJson != null) {
                context.sceneAdaptation = ExpressionSceneAdaptation.fromJson(sceneJson);
            }
            JSONObject emotionJson = json.optJSONObject("emotion_adaptation");
            if (emotionJson != null) {
                context.emotionAdaptation = ExpressionEmotionAdaptation.fromJson(emotionJson);
            }
            JSONObject socialJson = json.optJSONObject("social_adaptation");
            if (socialJson != null) {
                context.socialAdaptation = ExpressionSocialAdaptation.fromJson(socialJson);
            }
            JSONObject animJson = json.optJSONObject("animation_state");
            if (animJson != null) {
                context.animationState = AnimationState.fromJson(animJson);
            }
            return context;
        }
    }

    // === Expression State ===

    public static class ExpressionState {
        public String expressionName = "neutral";
        public double intensity = 0.5;
        public int duration = 500;
        public String transition = "smooth";

        public String eyebrowState = "neutral";
        public String eyeState = "normal";
        public String mouthState = "neutral";

        public boolean blendWithPrevious = false;
        public double blendProgress = 0.0;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("expression_name", expressionName);
                json.put("intensity", intensity);
                json.put("duration", duration);
                json.put("transition", transition);
                json.put("eyebrow_state", eyebrowState);
                json.put("eye_state", eyeState);
                json.put("mouth_state", mouthState);
                json.put("blend_with_previous", blendWithPrevious);
                json.put("blend_progress", blendProgress);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ExpressionState fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionState state = new ExpressionState();
            state.expressionName = json.optString("expression_name", "neutral");
            state.intensity = json.optDouble("intensity", 0.5);
            state.duration = json.optInt("duration", 500);
            state.transition = json.optString("transition", "smooth");
            state.eyebrowState = json.optString("eyebrow_state", "neutral");
            state.eyeState = json.optString("eye_state", "normal");
            state.mouthState = json.optString("mouth_state", "neutral");
            state.blendWithPrevious = json.optBoolean("blend_with_previous", false);
            state.blendProgress = json.optDouble("blend_progress", 0.0);
            return state;
        }
    }

    // === Gesture State ===

    public static class GestureState {
        public String gestureName = "neutral";
        public double intensity = 0.5;
        public int duration = 500;
        public String transition = "smooth";

        public String posture = "neutral";
        public String handPosition = "natural";
        public String headPosition = "neutral";

        public boolean isMirroring = false;
        public String mirroredFrom = "";
        public int mirroringDelay = 0;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("gesture_name", gestureName);
                json.put("intensity", intensity);
                json.put("duration", duration);
                json.put("transition", transition);
                json.put("posture", posture);
                json.put("hand_position", handPosition);
                json.put("head_position", headPosition);
                json.put("is_mirroring", isMirroring);
                json.put("mirrored_from", mirroredFrom);
                json.put("mirroring_delay", mirroringDelay);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static GestureState fromJson(@NonNull JSONObject json) throws JSONException {
            GestureState state = new GestureState();
            state.gestureName = json.optString("gesture_name", "neutral");
            state.intensity = json.optDouble("intensity", 0.5);
            state.duration = json.optInt("duration", 500);
            state.transition = json.optString("transition", "smooth");
            state.posture = json.optString("posture", "neutral");
            state.handPosition = json.optString("hand_position", "natural");
            state.headPosition = json.optString("head_position", "neutral");
            state.isMirroring = json.optBoolean("is_mirroring", false);
            state.mirroredFrom = json.optString("mirrored_from", "");
            state.mirroringDelay = json.optInt("mirroring_delay", 0);
            return state;
        }
    }

    // === Expression Scene Adaptation ===

    public static class ExpressionSceneAdaptation {
        public String scene = "default";
        public double expressionRange = 0.5;
        public double gestureRange = 0.5;
        public double formalityLevel = 0.5;
        public double eyeContactLevel = 0.5;
        public String idleAnimationStyle = "subtle";

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("scene", scene);
                json.put("expression_range", expressionRange);
                json.put("gesture_range", gestureRange);
                json.put("formality_level", formalityLevel);
                json.put("eye_contact_level", eyeContactLevel);
                json.put("idle_animation_style", idleAnimationStyle);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ExpressionSceneAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionSceneAdaptation adaptation = new ExpressionSceneAdaptation();
            adaptation.scene = json.optString("scene", "default");
            adaptation.expressionRange = json.optDouble("expression_range", 0.5);
            adaptation.gestureRange = json.optDouble("gesture_range", 0.5);
            adaptation.formalityLevel = json.optDouble("formality_level", 0.5);
            adaptation.eyeContactLevel = json.optDouble("eye_contact_level", 0.5);
            adaptation.idleAnimationStyle = json.optString("idle_animation_style", "subtle");
            return adaptation;
        }
    }

    // === Expression Emotion Adaptation ===

    public static class ExpressionEmotionAdaptation {
        public String currentEmotion = "neutral";
        public String targetExpression = "neutral";
        public String targetGesture = "neutral";
        public String transitionSpeed = "smooth";
        public double intensityMultiplier = 1.0;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("current_emotion", currentEmotion);
                json.put("target_expression", targetExpression);
                json.put("target_gesture", targetGesture);
                json.put("transition_speed", transitionSpeed);
                json.put("intensity_multiplier", intensityMultiplier);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ExpressionEmotionAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionEmotionAdaptation adaptation = new ExpressionEmotionAdaptation();
            adaptation.currentEmotion = json.optString("current_emotion", "neutral");
            adaptation.targetExpression = json.optString("target_expression", "neutral");
            adaptation.targetGesture = json.optString("target_gesture", "neutral");
            adaptation.transitionSpeed = json.optString("transition_speed", "smooth");
            adaptation.intensityMultiplier = json.optDouble("intensity_multiplier", 1.0);
            return adaptation;
        }
    }

    // === Expression Social Adaptation ===

    public static class ExpressionSocialAdaptation {
        public String socialContext = "casual";
        public String greetingGesture = "nod";
        public String farewellGesture = "nod";
        public double eyeContactTendency = 0.5;
        public boolean mirroringEnabled = false;
        public String touchPermission = "none";

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("social_context", socialContext);
                json.put("greeting_gesture", greetingGesture);
                json.put("farewell_gesture", farewellGesture);
                json.put("eye_contact_tendency", eyeContactTendency);
                json.put("mirroring_enabled", mirroringEnabled);
                json.put("touch_permission", touchPermission);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ExpressionSocialAdaptation fromJson(@NonNull JSONObject json) throws JSONException {
            ExpressionSocialAdaptation adaptation = new ExpressionSocialAdaptation();
            adaptation.socialContext = json.optString("social_context", "casual");
            adaptation.greetingGesture = json.optString("greeting_gesture", "nod");
            adaptation.farewellGesture = json.optString("farewell_gesture", "nod");
            adaptation.eyeContactTendency = json.optDouble("eye_contact_tendency", 0.5);
            adaptation.mirroringEnabled = json.optBoolean("mirroring_enabled", false);
            adaptation.touchPermission = json.optString("touch_permission", "none");
            return adaptation;
        }
    }

    // === Animation State ===

    public static class AnimationState {
        public String currentAnimation = "idle";
        public double animationProgress = 0.0;
        public int animationFPS = 30;
        public boolean isLooping = true;
        public boolean isBlending = false;
        public double blendProgress = 0.0;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("current_animation", currentAnimation);
                json.put("animation_progress", animationProgress);
                json.put("animation_fps", animationFPS);
                json.put("is_looping", isLooping);
                json.put("is_blending", isBlending);
                json.put("blend_progress", blendProgress);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static AnimationState fromJson(@NonNull JSONObject json) throws JSONException {
            AnimationState state = new AnimationState();
            state.currentAnimation = json.optString("current_animation", "idle");
            state.animationProgress = json.optDouble("animation_progress", 0.0);
            state.animationFPS = json.optInt("animation_fps", 30);
            state.isLooping = json.optBoolean("is_looping", true);
            state.isBlending = json.optBoolean("is_blending", false);
            state.blendProgress = json.optDouble("blend_progress", 0.0);
            return state;
        }
    }

    // === JSON Serialization ===

    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("facial_expression_settings", facialExpressionSettings.toJson());
            json.put("body_gesture_settings", bodyGestureSettings.toJson());
            json.put("emotion_expression_mapping", emotionExpressionMapping.toJson());
            json.put("social_gesture_settings", socialGestureSettings.toJson());
            json.put("animation_settings", animationSettings.toJson());
            json.put("context", context.toJson());
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    @NonNull
    public static ExpressionGestureState fromJson(@NonNull JSONObject json) throws JSONException {
        ExpressionGestureState state = new ExpressionGestureState();
        JSONObject facialJson = json.optJSONObject("facial_expression_settings");
        if (facialJson != null) {
            state.facialExpressionSettings = FacialExpressionSettings.fromJson(facialJson);
        }
        JSONObject bodyJson = json.optJSONObject("body_gesture_settings");
        if (bodyJson != null) {
            state.bodyGestureSettings = BodyGestureSettings.fromJson(bodyJson);
        }
        JSONObject mappingJson = json.optJSONObject("emotion_expression_mapping");
        if (mappingJson != null) {
            state.emotionExpressionMapping = EmotionExpressionMapping.fromJson(mappingJson);
        }
        JSONObject socialJson = json.optJSONObject("social_gesture_settings");
        if (socialJson != null) {
            state.socialGestureSettings = SocialGestureSettings.fromJson(socialJson);
        }
        JSONObject animJson = json.optJSONObject("animation_settings");
        if (animJson != null) {
            state.animationSettings = AnimationSettings.fromJson(animJson);
        }
        JSONObject contextJson = json.optJSONObject("context");
        if (contextJson != null) {
            state.context = ExpressionGestureContext.fromJson(contextJson);
        }
        return state;
    }
}