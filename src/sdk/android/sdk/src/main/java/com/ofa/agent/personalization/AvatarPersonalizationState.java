package com.ofa.agent.personalization;

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
 * AvatarPersonalization状态模型 (v5.4.0)
 *
 * 端侧接收 Center 推送的形象个性化状态，用于个性化展示和管理。
 * 深层个性化管理在 Center 端 PersonalizationEngine 完成。
 */
public class AvatarPersonalizationState {

    // === Image Preferences ===
    private ImagePreferences imagePreferences;

    // === Image Evolution ===
    private ImageEvolution imageEvolution;

    // === Scene Adaptation Settings ===
    private SceneAdaptationSettings sceneAdaptationSettings;

    // === Style Management ===
    private StyleManagement styleManagement;

    // === Personalization Context ===
    private PersonalizationContext context;

    public AvatarPersonalizationState() {
        this.imagePreferences = new ImagePreferences();
        this.imageEvolution = new ImageEvolution();
        this.sceneAdaptationSettings = new SceneAdaptationSettings();
        this.styleManagement = new StyleManagement();
        this.context = new PersonalizationContext();
    }

    // === Getters ===

    @NonNull
    public ImagePreferences getImagePreferences() {
        return imagePreferences;
    }

    @NonNull
    public ImageEvolution getImageEvolution() {
        return imageEvolution;
    }

    @NonNull
    public SceneAdaptationSettings getSceneAdaptationSettings() {
        return sceneAdaptationSettings;
    }

    @NonNull
    public StyleManagement getStyleManagement() {
        return styleManagement;
    }

    @NonNull
    public PersonalizationContext getContext() {
        return context;
    }

    // === Image Preferences ===

    public static class ImagePreferences {
        public List<String> preferredColors = new ArrayList<>();
        public List<String> avoidedColors = new ArrayList<>();
        public String colorHarmonyStyle = "complementary";
        public String colorIntensity = "moderate";

        public List<String> preferredStyles = new ArrayList<>();
        public List<String> avoidedStyles = new ArrayList<>();
        public double styleMixingLevel = 0.5;
        public double styleExperimentation = 0.3;

        public String comfortPriority = "medium";
        public List<String> fabricPreferences = new ArrayList<>();
        public String fitPreferences = "fitted";
        public boolean weatherAdaptation = true;

        public List<String> favoriteBrands = new ArrayList<>();
        public List<String> avoidedBrands = new ArrayList<>();
        public double brandLoyalty = 0.5;
        public double localBrandSupport = 0.3;

        public String accessoryFrequency = "occasional";
        public List<String> accessoryTypes = new ArrayList<>();
        public String jewelryStyle = "minimal";
        public String watchStyle = "classic";

        public String groomingRoutine = "moderate";
        public String hairProductUse = "minimal";
        public String skincareRoutine = "basic";
        public String fragranceUse = "occasional";

        public String presentationEffort = "medium";
        public double attentionToDetail = 0.5;
        public double occasionAwareness = 0.7;

        public boolean isHighPresentationEffort() {
            return "high".equals(presentationEffort);
        }

        public boolean isExperimental() {
            return styleExperimentation > 0.6;
        }

        public boolean isComfortOriented() {
            return "high".equals(comfortPriority);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("preferred_colors", new JSONArray(preferredColors));
                json.put("avoided_colors", new JSONArray(avoidedColors));
                json.put("color_harmony_style", colorHarmonyStyle);
                json.put("color_intensity", colorIntensity);
                json.put("preferred_styles", new JSONArray(preferredStyles));
                json.put("avoided_styles", new JSONArray(avoidedStyles));
                json.put("style_mixing_level", styleMixingLevel);
                json.put("style_experimentation", styleExperimentation);
                json.put("comfort_priority", comfortPriority);
                json.put("fabric_preferences", new JSONArray(fabricPreferences));
                json.put("fit_preferences", fitPreferences);
                json.put("weather_adaptation", weatherAdaptation);
                json.put("favorite_brands", new JSONArray(favoriteBrands));
                json.put("avoided_brands", new JSONArray(avoidedBrands));
                json.put("brand_loyalty", brandLoyalty);
                json.put("local_brand_support", localBrandSupport);
                json.put("accessory_frequency", accessoryFrequency);
                json.put("accessory_types", new JSONArray(accessoryTypes));
                json.put("jewelry_style", jewelryStyle);
                json.put("watch_style", watchStyle);
                json.put("grooming_routine", groomingRoutine);
                json.put("hair_product_use", hairProductUse);
                json.put("skincare_routine", skincareRoutine);
                json.put("fragrance_use", fragranceUse);
                json.put("presentation_effort", presentationEffort);
                json.put("attention_to_detail", attentionToDetail);
                json.put("occasion_awareness", occasionAwareness);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ImagePreferences fromJson(@NonNull JSONObject json) throws JSONException {
            ImagePreferences prefs = new ImagePreferences();
            JSONArray prefColors = json.optJSONArray("preferred_colors");
            if (prefColors != null) {
                for (int i = 0; i < prefColors.length(); i++) {
                    prefs.preferredColors.add(prefColors.getString(i));
                }
            }
            JSONArray avoidColors = json.optJSONArray("avoided_colors");
            if (avoidColors != null) {
                for (int i = 0; i < avoidColors.length(); i++) {
                    prefs.avoidedColors.add(avoidColors.getString(i));
                }
            }
            prefs.colorHarmonyStyle = json.optString("color_harmony_style", "complementary");
            prefs.colorIntensity = json.optString("color_intensity", "moderate");
            JSONArray prefStyles = json.optJSONArray("preferred_styles");
            if (prefStyles != null) {
                for (int i = 0; i < prefStyles.length(); i++) {
                    prefs.preferredStyles.add(prefStyles.getString(i));
                }
            }
            JSONArray avoidStyles = json.optJSONArray("avoided_styles");
            if (avoidStyles != null) {
                for (int i = 0; i < avoidStyles.length(); i++) {
                    prefs.avoidedStyles.add(avoidStyles.getString(i));
                }
            }
            prefs.styleMixingLevel = json.optDouble("style_mixing_level", 0.5);
            prefs.styleExperimentation = json.optDouble("style_experimentation", 0.3);
            prefs.comfortPriority = json.optString("comfort_priority", "medium");
            JSONArray fabrics = json.optJSONArray("fabric_preferences");
            if (fabrics != null) {
                for (int i = 0; i < fabrics.length(); i++) {
                    prefs.fabricPreferences.add(fabrics.getString(i));
                }
            }
            prefs.fitPreferences = json.optString("fit_preferences", "fitted");
            prefs.weatherAdaptation = json.optBoolean("weather_adaptation", true);
            JSONArray favBrands = json.optJSONArray("favorite_brands");
            if (favBrands != null) {
                for (int i = 0; i < favBrands.length(); i++) {
                    prefs.favoriteBrands.add(favBrands.getString(i));
                }
            }
            JSONArray avoidBrands = json.optJSONArray("avoided_brands");
            if (avoidBrands != null) {
                for (int i = 0; i < avoidBrands.length(); i++) {
                    prefs.avoidedBrands.add(avoidBrands.getString(i));
                }
            }
            prefs.brandLoyalty = json.optDouble("brand_loyalty", 0.5);
            prefs.localBrandSupport = json.optDouble("local_brand_support", 0.3);
            prefs.accessoryFrequency = json.optString("accessory_frequency", "occasional");
            JSONArray accTypes = json.optJSONArray("accessory_types");
            if (accTypes != null) {
                for (int i = 0; i < accTypes.length(); i++) {
                    prefs.accessoryTypes.add(accTypes.getString(i));
                }
            }
            prefs.jewelryStyle = json.optString("jewelry_style", "minimal");
            prefs.watchStyle = json.optString("watch_style", "classic");
            prefs.groomingRoutine = json.optString("grooming_routine", "moderate");
            prefs.hairProductUse = json.optString("hair_product_use", "minimal");
            prefs.skincareRoutine = json.optString("skincare_routine", "basic");
            prefs.fragranceUse = json.optString("fragrance_use", "occasional");
            prefs.presentationEffort = json.optString("presentation_effort", "medium");
            prefs.attentionToDetail = json.optDouble("attention_to_detail", 0.5);
            prefs.occasionAwareness = json.optDouble("occasion_awareness", 0.7);
            return prefs;
        }
    }

    // === Image Evolution ===

    public static class ImageEvolution {
        public String evolutionMode = "gradual";
        public String evolutionSpeed = "moderate";
        public String evolutionTendency = "balanced";

        public List<EvolutionTrigger> evolutionTriggers = new ArrayList<>();
        public List<StyleRecord> styleHistory = new ArrayList<>();
        public List<StyleMilestone> styleMilestones = new ArrayList<>();

        public List<String> coreStyleElements = new ArrayList<>();
        public List<String> flexibleElements = new ArrayList<>();
        public List<String> experimentalZone = new ArrayList<>();

        public boolean seasonalAdaptation = true;
        public String springStyle = "fresh";
        public String summerStyle = "light";
        public String autumnStyle = "warm";
        public String winterStyle = "cozy";

        public Map<String, StageStyleRule> lifeStageStyleRules = new HashMap<>();

        public double trendFollowingLevel = 0.4;
        public String trendAdoptionSpeed = "mainstream";
        public String trendFilter = "selective";

        public boolean isStable() {
            return "stable".equals(evolutionMode);
        }

        public boolean isExperimental() {
            return "experimental".equals(evolutionMode);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("evolution_mode", evolutionMode);
                json.put("evolution_speed", evolutionSpeed);
                json.put("evolution_tendency", evolutionTendency);
                json.put("seasonal_adaptation", seasonalAdaptation);
                json.put("spring_style", springStyle);
                json.put("summer_style", summerStyle);
                json.put("autumn_style", autumnStyle);
                json.put("winter_style", winterStyle);
                json.put("trend_following_level", trendFollowingLevel);
                json.put("trend_adoption_speed", trendAdoptionSpeed);
                json.put("trend_filter", trendFilter);
                json.put("core_style_elements", new JSONArray(coreStyleElements));
                json.put("flexible_elements", new JSONArray(flexibleElements));
                json.put("experimental_zone", new JSONArray(experimentalZone));
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static ImageEvolution fromJson(@NonNull JSONObject json) throws JSONException {
            ImageEvolution evolution = new ImageEvolution();
            evolution.evolutionMode = json.optString("evolution_mode", "gradual");
            evolution.evolutionSpeed = json.optString("evolution_speed", "moderate");
            evolution.evolutionTendency = json.optString("evolution_tendency", "balanced");
            evolution.seasonalAdaptation = json.optBoolean("seasonal_adaptation", true);
            evolution.springStyle = json.optString("spring_style", "fresh");
            evolution.summerStyle = json.optString("summer_style", "light");
            evolution.autumnStyle = json.optString("autumn_style", "warm");
            evolution.winterStyle = json.optString("winter_style", "cozy");
            evolution.trendFollowingLevel = json.optDouble("trend_following_level", 0.4);
            evolution.trendAdoptionSpeed = json.optString("trend_adoption_speed", "mainstream");
            evolution.trendFilter = json.optString("trend_filter", "selective");
            JSONArray coreElements = json.optJSONArray("core_style_elements");
            if (coreElements != null) {
                for (int i = 0; i < coreElements.length(); i++) {
                    evolution.coreStyleElements.add(coreElements.getString(i));
                }
            }
            JSONArray flexElements = json.optJSONArray("flexible_elements");
            if (flexElements != null) {
                for (int i = 0; i < flexElements.length(); i++) {
                    evolution.flexibleElements.add(flexElements.getString(i));
                }
            }
            JSONArray expElements = json.optJSONArray("experimental_zone");
            if (expElements != null) {
                for (int i = 0; i < expElements.length(); i++) {
                    evolution.experimentalZone.add(expElements.getString(i));
                }
            }
            return evolution;
        }
    }

    // === Evolution Trigger ===

    public static class EvolutionTrigger {
        public String triggerType;
        public String triggerName;
        public String styleChange;
        public int priority;
        public boolean autoApply;
    }

    // === Style Record ===

    public static class StyleRecord {
        public long timestamp;
        public String styleName;
        public String description;
        public String context;
        public double satisfaction;
    }

    // === Style Milestone ===

    public static class StyleMilestone {
        public long date;
        public String milestoneName;
        public String description;
        public String beforeStyle;
        public String afterStyle;
        public double significance;
    }

    // === Stage Style Rule ===

    public static class StageStyleRule {
        public String stageName;
        public List<String> allowedStyles = new ArrayList<>();
        public List<String> forbiddenStyles = new ArrayList<>();
        public String styleTone;
        public String complexityLevel;
    }

    // === Scene Adaptation Settings ===

    public static class SceneAdaptationSettings {
        public String adaptationMode = "auto";
        public String adaptationSpeed = "gradual";
        public String adaptationIntensity = "moderate";

        public List<SceneRule> sceneRules = new ArrayList<>();

        public String defaultWorkStyle = "business";
        public String defaultHomeStyle = "casual";
        public String defaultSocialStyle = "smart_casual";
        public String defaultActiveStyle = "sporty";

        public String transitionStyle = "gradual";
        public int transitionDuration = 5;
        public String transitionAnimation = "fade";

        public boolean locationAwareness = true;
        public boolean timeAwareness = true;
        public boolean calendarAwareness = true;
        public boolean weatherAwareness = true;

        public String privacyMode = "semi_private";
        public double stylePrivacyLevel = 0.5;

        public boolean isAutoAdaptation() {
            return "auto".equals(adaptationMode);
        }

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("adaptation_mode", adaptationMode);
                json.put("adaptation_speed", adaptationSpeed);
                json.put("adaptation_intensity", adaptationIntensity);
                json.put("default_work_style", defaultWorkStyle);
                json.put("default_home_style", defaultHomeStyle);
                json.put("default_social_style", defaultSocialStyle);
                json.put("default_active_style", defaultActiveStyle);
                json.put("transition_style", transitionStyle);
                json.put("transition_duration", transitionDuration);
                json.put("transition_animation", transitionAnimation);
                json.put("location_awareness", locationAwareness);
                json.put("time_awareness", timeAwareness);
                json.put("calendar_awareness", calendarAwareness);
                json.put("weather_awareness", weatherAwareness);
                json.put("privacy_mode", privacyMode);
                json.put("style_privacy_level", stylePrivacyLevel);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static SceneAdaptationSettings fromJson(@NonNull JSONObject json) throws JSONException {
            SceneAdaptationSettings settings = new SceneAdaptationSettings();
            settings.adaptationMode = json.optString("adaptation_mode", "auto");
            settings.adaptationSpeed = json.optString("adaptation_speed", "gradual");
            settings.adaptationIntensity = json.optString("adaptation_intensity", "moderate");
            settings.defaultWorkStyle = json.optString("default_work_style", "business");
            settings.defaultHomeStyle = json.optString("default_home_style", "casual");
            settings.defaultSocialStyle = json.optString("default_social_style", "smart_casual");
            settings.defaultActiveStyle = json.optString("default_active_style", "sporty");
            settings.transitionStyle = json.optString("transition_style", "gradual");
            settings.transitionDuration = json.optInt("transition_duration", 5);
            settings.transitionAnimation = json.optString("transition_animation", "fade");
            settings.locationAwareness = json.optBoolean("location_awareness", true);
            settings.timeAwareness = json.optBoolean("time_awareness", true);
            settings.calendarAwareness = json.optBoolean("calendar_awareness", true);
            settings.weatherAwareness = json.optBoolean("weather_awareness", true);
            settings.privacyMode = json.optString("privacy_mode", "semi_private");
            settings.stylePrivacyLevel = json.optDouble("style_privacy_level", 0.5);
            return settings;
        }
    }

    // === Scene Rule ===

    public static class SceneRule {
        public String sceneName;
        public String sceneType;
        public String requiredStyle;
        public List<String> allowedVariations = new ArrayList<>();
        public List<String> forbiddenElements = new ArrayList<>();
        public String accessoryRule;
        public String groomingRule;
        public String expressionRule;
        public String postureRule;
        public int priority;
        public boolean enabled;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("scene_name", sceneName);
                json.put("scene_type", sceneType);
                json.put("required_style", requiredStyle);
                json.put("accessory_rule", accessoryRule);
                json.put("grooming_rule", groomingRule);
                json.put("expression_rule", expressionRule);
                json.put("posture_rule", postureRule);
                json.put("priority", priority);
                json.put("enabled", enabled);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static SceneRule fromJson(@NonNull JSONObject json) throws JSONException {
            SceneRule rule = new SceneRule();
            rule.sceneName = json.optString("scene_name", "");
            rule.sceneType = json.optString("scene_type", "");
            rule.requiredStyle = json.optString("required_style", "casual");
            rule.accessoryRule = json.optString("accessory_rule", "minimal");
            rule.groomingRule = json.optString("grooming_rule", "natural");
            rule.expressionRule = json.optString("expression_rule", "neutral");
            rule.postureRule = json.optString("posture_rule", "relaxed");
            rule.priority = json.optInt("priority", 5);
            rule.enabled = json.optBoolean("enabled", true);
            return rule;
        }
    }

    // === Style Management ===

    public static class StyleManagement {
        public List<StyleCollection> styleCollections = new ArrayList<>();
        public List<OutfitRecord> favoriteOutfits = new ArrayList<>();
        public List<OutfitRecord> outfitHistory = new ArrayList<>();

        public List<StyleTemplate> styleTemplates = new ArrayList<>();
        public String defaultTemplate;

        public boolean recommendationEnabled = true;
        public String recommendationSource = "hybrid";
        public String recommendationFrequency = "weekly";

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("recommendation_enabled", recommendationEnabled);
                json.put("recommendation_source", recommendationSource);
                json.put("recommendation_frequency", recommendationFrequency);
                json.put("default_template", defaultTemplate);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static StyleManagement fromJson(@NonNull JSONObject json) throws JSONException {
            StyleManagement mgmt = new StyleManagement();
            mgmt.recommendationEnabled = json.optBoolean("recommendation_enabled", true);
            mgmt.recommendationSource = json.optString("recommendation_source", "hybrid");
            mgmt.recommendationFrequency = json.optString("recommendation_frequency", "weekly");
            mgmt.defaultTemplate = json.optString("default_template", "");
            return mgmt;
        }
    }

    // === Style Collection ===

    public static class StyleCollection {
        public String collectionId;
        public String collectionName;
        public String description;
        public List<String> styles = new ArrayList<>();
        public List<String> occasions = new ArrayList<>();
        public List<String> seasons = new ArrayList<>();
        public boolean isActive;
        public long createdAt;
    }

    // === Outfit Record ===

    public static class OutfitRecord {
        public String outfitId;
        public String outfitName;
        public String description;
        public List<String> items = new ArrayList<>();
        public String styleCategory;
        public String occasion;
        public String season;
        public double rating;
        public int wearCount;
        public long lastWorn;
        public boolean isFavorite;
    }

    // === Style Template ===

    public static class StyleTemplate {
        public String templateId;
        public String templateName;
        public String description;
        public String styleBase;
        public List<String> variations = new ArrayList<>();
        public List<String> requiredItems = new ArrayList<>();
        public List<String> optionalItems = new ArrayList<>();
        public List<String> colorScheme = new ArrayList<>();
        public boolean isDefault;
    }

    // === Wardrobe Item ===

    public static class WardrobeItem {
        public String itemId;
        public String itemName;
        public String category;
        public String subcategory;
        public String color;
        public String pattern;
        public String style;
        public String brand;
        public String size;
        public String material;
        public String season;
        public List<String> occasions = new ArrayList<>();
        public int wearCount;
        public String condition;
        public boolean isFavorite;
        public boolean isActive;
    }

    // === Wardrobe Stats ===

    public static class WardrobeStats {
        public int totalItems;
        public Map<String, Integer> byCategory = new HashMap<>();
        public Map<String, Integer> byColor = new HashMap<>();
        public Map<String, Integer> bySeason = new HashMap<>();
        public Map<String, Integer> byStyle = new HashMap<>();
        public double averageWearCount;
        public int unusedItems;
        public int favoriteCount;
        public double totalValue;
    }

    // === Personalization Context ===

    public static class PersonalizationContext {
        public String identityId;

        public StyleRecommendation recommendedStyle = new StyleRecommendation();
        public OutfitRecommendation recommendedOutfit = new OutfitRecommendation();
        public List<String> recommendedAccessories = new ArrayList<>();

        public String currentScene;
        public double sceneStyleMatch;
        public boolean sceneAdaptationNeeded;

        public double styleScore;
        public double consistencyScore;
        public double versatilityScore;
        public double authenticityScore;

        public String evolutionStage;
        public String nextEvolutionPreview;
        public double evolutionReadiness;

        public List<StyleSuggestion> styleSuggestions = new ArrayList<>();
        public List<String> improvementTips = new ArrayList<>();
        public List<TrendAlert> trendAlerts = new ArrayList<>();

        public long timestamp;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("identity_id", identityId);
                json.put("recommended_style", recommendedStyle.toJson());
                json.put("current_scene", currentScene);
                json.put("scene_style_match", sceneStyleMatch);
                json.put("scene_adaptation_needed", sceneAdaptationNeeded);
                json.put("style_score", styleScore);
                json.put("consistency_score", consistencyScore);
                json.put("versatility_score", versatilityScore);
                json.put("authenticity_score", authenticityScore);
                json.put("evolution_stage", evolutionStage);
                json.put("next_evolution_preview", nextEvolutionPreview);
                json.put("evolution_readiness", evolutionReadiness);
                json.put("timestamp", timestamp);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static PersonalizationContext fromJson(@NonNull JSONObject json) throws JSONException {
            PersonalizationContext context = new PersonalizationContext();
            context.identityId = json.optString("identity_id", "");
            context.currentScene = json.optString("current_scene", "");
            context.sceneStyleMatch = json.optDouble("scene_style_match", 0.5);
            context.sceneAdaptationNeeded = json.optBoolean("scene_adaptation_needed", false);
            context.styleScore = json.optDouble("style_score", 0.5);
            context.consistencyScore = json.optDouble("consistency_score", 0.5);
            context.versatilityScore = json.optDouble("versatility_score", 0.5);
            context.authenticityScore = json.optDouble("authenticity_score", 0.5);
            context.evolutionStage = json.optString("evolution_stage", "gradual");
            context.nextEvolutionPreview = json.optString("next_evolution_preview", "");
            context.evolutionReadiness = json.optDouble("evolution_readiness", 0.5);
            context.timestamp = json.optLong("timestamp", System.currentTimeMillis());
            JSONObject recStyle = json.optJSONObject("recommended_style");
            if (recStyle != null) {
                context.recommendedStyle = StyleRecommendation.fromJson(recStyle);
            }
            return context;
        }
    }

    // === Style Recommendation ===

    public static class StyleRecommendation {
        public String styleName;
        public double confidence;
        public String reason;
        public List<String> colorPalette = new ArrayList<>();
        public List<String> keyPieces = new ArrayList<>();
        public String occasion;
        public String season;
        public String weather;
        public double vibeMatch;

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("style_name", styleName);
                json.put("confidence", confidence);
                json.put("reason", reason);
                json.put("color_palette", new JSONArray(colorPalette));
                json.put("occasion", occasion);
                json.put("season", season);
                json.put("vibe_match", vibeMatch);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static StyleRecommendation fromJson(@NonNull JSONObject json) throws JSONException {
            StyleRecommendation rec = new StyleRecommendation();
            rec.styleName = json.optString("style_name", "");
            rec.confidence = json.optDouble("confidence", 0.5);
            rec.reason = json.optString("reason", "");
            rec.occasion = json.optString("occasion", "");
            rec.season = json.optString("season", "");
            rec.vibeMatch = json.optDouble("vibe_match", 0.5);
            JSONArray colors = json.optJSONArray("color_palette");
            if (colors != null) {
                for (int i = 0; i < colors.length(); i++) {
                    rec.colorPalette.add(colors.getString(i));
                }
            }
            return rec;
        }
    }

    // === Outfit Recommendation ===

    public static class OutfitRecommendation {
        public String outfitId;
        public String outfitName;
        public List<String> items = new ArrayList<>();
        public String styleCategory;
        public String occasion;
        public String weather;
        public double confidence;
        public String reason;
        public List<String> alternatives = new ArrayList<>();

        @NonNull
        public JSONObject toJson() {
            JSONObject json = new JSONObject();
            try {
                json.put("outfit_id", outfitId);
                json.put("outfit_name", outfitName);
                json.put("style_category", styleCategory);
                json.put("occasion", occasion);
                json.put("confidence", confidence);
                json.put("reason", reason);
            } catch (JSONException e) {
                // ignore
            }
            return json;
        }

        public static OutfitRecommendation fromJson(@NonNull JSONObject json) throws JSONException {
            OutfitRecommendation rec = new OutfitRecommendation();
            rec.outfitId = json.optString("outfit_id", "");
            rec.outfitName = json.optString("outfit_name", "");
            rec.styleCategory = json.optString("style_category", "");
            rec.occasion = json.optString("occasion", "");
            rec.confidence = json.optDouble("confidence", 0.5);
            rec.reason = json.optString("reason", "");
            return rec;
        }
    }

    // === Style Suggestion ===

    public static class StyleSuggestion {
        public String suggestionType;
        public String targetArea;
        public String suggestion;
        public int priority;
        public String effort;
        public double impact;
    }

    // === Trend Alert ===

    public static class TrendAlert {
        public String trendName;
        public String description;
        public double relevance;
        public String adoptionLevel;
        public String category;
        public List<String> suggestedItems = new ArrayList<>();
        public String howToAdopt;
        public String riskLevel;
    }

    // === JSON Serialization ===

    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("image_preferences", imagePreferences.toJson());
            json.put("image_evolution", imageEvolution.toJson());
            json.put("scene_adaptation_settings", sceneAdaptationSettings.toJson());
            json.put("style_management", styleManagement.toJson());
            json.put("context", context.toJson());
        } catch (JSONException e) {
            // ignore
        }
        return json;
    }

    @NonNull
    public static AvatarPersonalizationState fromJson(@NonNull JSONObject json) throws JSONException {
        AvatarPersonalizationState state = new AvatarPersonalizationState();
        JSONObject prefsJson = json.optJSONObject("image_preferences");
        if (prefsJson != null) {
            state.imagePreferences = ImagePreferences.fromJson(prefsJson);
        }
        JSONObject evolutionJson = json.optJSONObject("image_evolution");
        if (evolutionJson != null) {
            state.imageEvolution = ImageEvolution.fromJson(evolutionJson);
        }
        JSONObject sceneJson = json.optJSONObject("scene_adaptation_settings");
        if (sceneJson != null) {
            state.sceneAdaptationSettings = SceneAdaptationSettings.fromJson(sceneJson);
        }
        JSONObject mgmtJson = json.optJSONObject("style_management");
        if (mgmtJson != null) {
            state.styleManagement = StyleManagement.fromJson(mgmtJson);
        }
        JSONObject contextJson = json.optJSONObject("context");
        if (contextJson != null) {
            state.context = PersonalizationContext.fromJson(contextJson);
        }
        return state;
    }
}