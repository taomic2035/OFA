package com.ofa.agent.ai;

import android.content.Context;
import android.graphics.Bitmap;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.ai.decision.SmartDecisionEngine;
import com.ofa.agent.ai.intent.LocalIntentClassifier;
import com.ofa.agent.ai.recommendation.OperationRecommender;
import com.ofa.agent.ai.vision.UIElementRecognizer;
import com.ofa.agent.automation.AutomationOrchestrator;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.intent.UserIntent;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.skill.SkillDefinition;

import org.json.JSONObject;

import java.util.List;
import java.util.Map;

/**
 * AI-Enhanced Automation Orchestrator - combines automation with AI capabilities.
 * Provides intelligent intent recognition, decision making, and recommendations.
 */
public class AIEnhancedOrchestrator {

    private static final String TAG = "AIEnhancedOrchestrator";

    private final Context context;
    private final AutomationOrchestrator baseOrchestrator;

    // AI components
    private LocalAIEngine aiEngine;
    private LocalIntentClassifier intentClassifier;
    private SmartDecisionEngine decisionEngine;
    private OperationRecommender operationRecommender;
    private UIElementRecognizer elementRecognizer;

    private boolean aiEnabled = true;
    private boolean initialized = false;

    /**
     * Create AI-enhanced orchestrator
     */
    public AIEnhancedOrchestrator(@NonNull Context context,
                                   @Nullable UserMemoryManager memoryManager) {
        this.context = context;
        this.baseOrchestrator = new AutomationOrchestrator(context);

        // Initialize AI components
        initializeAI(memoryManager);
    }

    /**
     * Initialize AI components
     */
    private void initializeAI(@Nullable UserMemoryManager memoryManager) {
        Log.i(TAG, "Initializing AI components...");

        // AI Engine
        aiEngine = new LocalAIEngine(context);
        aiEngine.initialize(InferenceConfig.standard());

        // Intent Classifier
        intentClassifier = new LocalIntentClassifier(context);
        intentClassifier.initialize();

        // Decision Engine
        decisionEngine = new SmartDecisionEngine(context, memoryManager);

        // Operation Recommender
        operationRecommender = new OperationRecommender(context, memoryManager,
            baseOrchestrator.getAdapterManager());

        // Element Recognizer
        elementRecognizer = new UIElementRecognizer(context);
        elementRecognizer.initialize();

        Log.i(TAG, "AI components initialized");
    }

    /**
     * Initialize the orchestrator
     */
    public void initialize() {
        baseOrchestrator.initialize();
        initialized = true;

        Log.i(TAG, "AI-Enhanced Orchestrator initialized");
    }

    /**
     * Enable/disable AI features
     */
    public void setAIEnabled(boolean enabled) {
        this.aiEnabled = enabled;
        Log.i(TAG, "AI features " + (enabled ? "enabled" : "disabled"));
    }

    // ===== Natural Language Interface =====

    /**
     * Process natural language input and execute
     */
    @NonNull
    public AutomationResult processNaturalLanguage(@NonNull String input) {
        Log.i(TAG, "Processing: " + input);

        // Classify intent
        UserIntent intent = intentClassifier.classify(input);

        if (intent == null) {
            Log.w(TAG, "Could not classify intent for: " + input);
            return new AutomationResult("nlp", "Could not understand: " + input);
        }

        Log.i(TAG, "Recognized intent: " + intent.getIntentName() +
            " (confidence: " + intent.getConfidence() + ")");

        // Execute based on intent
        return executeFromIntent(intent);
    }

    /**
     * Execute automation from intent
     */
    @NonNull
    private AutomationResult executeFromIntent(@NonNull UserIntent intent) {
        String intentName = intent.getIntentName();
        Map<String, String> slots = intent.getSlots();

        switch (intentName) {
            case "search":
                return baseOrchestrator.execute("search", slots);

            case "app_launch":
                return baseOrchestrator.execute("app_launch", slots);

            case "order_food":
                return executeFoodOrder(slots);

            case "shopping":
                return executeShopping(slots);

            default:
                return baseOrchestrator.execute(intentName, slots);
        }
    }

    /**
     * Execute food order with AI assistance
     */
    @NonNull
    private AutomationResult executeFoodOrder(@NonNull Map<String, String> slots) {
        // Use decision engine to select best options
        String foodType = slots.getOrDefault("food_type", "外卖");

        // Get recommendations
        List<OperationRecommender.Recommendation> recommendations =
            operationRecommender.getRecommendations(Map.of("intent", "order_food", "food_type", foodType));

        Log.i(TAG, "Got " + recommendations.size() + " recommendations for food order");

        // Execute template
        return baseOrchestrator.executeTemplate("food_delivery", slots);
    }

    /**
     * Execute shopping with AI assistance
     */
    @NonNull
    private AutomationResult executeShopping(@NonNull Map<String, String> slots) {
        return baseOrchestrator.executeTemplate("shopping", slots);
    }

    // ===== Smart Operations =====

    /**
     * Execute with AI-optimized decisions
     */
    @NonNull
    public AutomationResult executeSmart(@NonNull String operation,
                                          @NonNull Map<String, String> params) {
        // Get recommendations
        List<OperationRecommender.Recommendation> recommendations =
            operationRecommender.getRecommendations(Map.of("operation", operation));

        // If AI suggests different operation, log it
        if (!recommendations.isEmpty() &&
            !recommendations.get(0).operation.equals(operation)) {
            Log.d(TAG, "AI suggests: " + recommendations.get(0).operation +
                " instead of " + operation);
        }

        // Execute operation
        long startTime = System.currentTimeMillis();
        AutomationResult result = baseOrchestrator.execute(operation, params);
        long duration = System.currentTimeMillis() - startTime;

        // Record for learning
        operationRecommender.recordOperation(operation, result.isSuccess(), duration);

        // Update decision engine
        double reward = decisionEngine.calculateReward(result);
        decisionEngine.reportReward(SmartDecisionEngine.DECISION_TIMING, operation, reward);

        return result;
    }

    /**
     * Get operation recommendations
     */
    @NonNull
    public List<OperationRecommender.Recommendation> getRecommendations() {
        String page = baseOrchestrator.getCurrentPage();
        return operationRecommender.getRecommendations(Map.of("page", page));
    }

    /**
     * Get intent alternatives
     */
    @NonNull
    public List<UserIntent> getIntentAlternatives(@NonNull String input, int topK) {
        return intentClassifier.classifyWithAlternatives(input, topK);
    }

    // ===== Vision Operations =====

    /**
     * Analyze screenshot
     */
    @NonNull
    public List<UIElementRecognizer.RecognizedElement> analyzeScreen(@NonNull Bitmap screenshot) {
        return elementRecognizer.recognize(screenshot);
    }

    /**
     * Find element by text in screenshot
     */
    @Nullable
    public UIElementRecognizer.RecognizedElement findElementByText(@NonNull Bitmap screenshot,
                                                                    @NonNull String text) {
        List<UIElementRecognizer.RecognizedElement> elements = elementRecognizer.recognize(screenshot);
        return elementRecognizer.findContainingText(elements, text);
    }

    // ===== Decision Support =====

    /**
     * Get optimal shop choice
     */
    @Nullable
    public String getOptimalShop(@NonNull String category, @NonNull List<String> options) {
        decisionEngine.registerOptions(SmartDecisionEngine.DECISION_SHOP_SELECTION, options);
        return decisionEngine.selectOption(SmartDecisionEngine.DECISION_SHOP_SELECTION);
    }

    /**
     * Get optimal payment method
     */
    @Nullable
    public String getOptimalPaymentMethod() {
        return decisionEngine.getBestOption(SmartDecisionEngine.DECISION_PAYMENT_METHOD);
    }

    /**
     * Report decision outcome
     */
    public void reportDecisionOutcome(@NonNull String decisionType,
                                        @NonNull String option,
                                        double reward) {
        decisionEngine.reportReward(decisionType, option, reward);
    }

    // ===== Statistics & Status =====

    /**
     * Check if AI is ready
     */
    public boolean isAIReady() {
        return aiEnabled &&
            intentClassifier.isReady() &&
            aiEngine.isReady() &&
            elementRecognizer.isReady();
    }

    /**
     * Get AI status
     */
    @NonNull
    public String getAIStatus() {
        StringBuilder sb = new StringBuilder();
        sb.append("AI Status:\n");
        sb.append("  AI Engine: ").append(aiEngine.isReady() ? "Ready" : "Not Ready").append("\n");
        sb.append("  Intent Classifier: ").append(intentClassifier.isReady() ? "Ready" : "Not Ready").append("\n");
        sb.append("  Element Recognizer: ").append(elementRecognizer.isReady() ? "Ready" : "Not Ready").append("\n");
        sb.append("  Decision Engine: ").append("Active").append("\n");
        sb.append("  Loaded Models: ").append(aiEngine.getLoadedModels().length).append("\n");
        return sb.toString();
    }

    /**
     * Get full status report
     */
    @NonNull
    public String getFullStatusReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== AI-Enhanced Automation Orchestrator ===\n\n");

        sb.append("Base Orchestrator:\n");
        sb.append(baseOrchestrator.getStatusReport()).append("\n\n");

        sb.append("AI Components:\n");
        sb.append(getAIStatus()).append("\n");

        sb.append("Decision Stats:\n");
        for (String type : decisionEngine.getDecisionTypes()) {
            sb.append("  ").append(decisionEngine.getDecisionStats(type)).append("\n");
        }

        return sb.toString();
    }

    // ===== Delegates to Base Orchestrator =====

    @NonNull
    public AutomationResult execute(@NonNull String operation, @NonNull Map<String, String> params) {
        return baseOrchestrator.execute(operation, params);
    }

    @NonNull
    public AutomationResult executeSkill(@NonNull SkillDefinition skill, @NonNull Map<String, String> params) {
        return baseOrchestrator.executeSkill(skill, params);
    }

    @NonNull
    public AutomationResult executeTemplate(@NonNull String templateId, @NonNull Map<String, String> params) {
        return baseOrchestrator.executeTemplate(templateId, params);
    }

    public boolean isInitialized() {
        return initialized;
    }

    public boolean isAvailable() {
        return baseOrchestrator.isAvailable();
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        Log.i(TAG, "Shutting down AI-Enhanced Orchestrator...");

        baseOrchestrator.shutdown();
        intentClassifier.shutdown();
        elementRecognizer.shutdown();
        aiEngine.shutdown();

        initialized = false;
        Log.i(TAG, "Shutdown complete");
    }
}