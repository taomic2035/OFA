package com.ofa.agent.automation;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.adapter.AppAdapter;
import com.ofa.agent.automation.adapter.AppAdapterManager;
import com.ofa.agent.automation.hybrid.HybridAutomationEngine;
import com.ofa.agent.automation.integration.MemoryAwareAutomation;
import com.ofa.agent.automation.integration.IntentTriggeredAutomation;
import com.ofa.agent.automation.integration.SkillAutomationBridge;
import com.ofa.agent.automation.monitor.AutomationLogger;
import com.ofa.agent.automation.monitor.PerformanceMonitor;
import com.ofa.agent.automation.recovery.ErrorRecovery;
import com.ofa.agent.automation.recovery.RetryPolicy;
import com.ofa.agent.automation.template.TemplateRegistry;
import com.ofa.agent.intent.IntentEngine;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.skill.SkillDefinition;

import org.json.JSONObject;

import java.util.Map;

/**
 * Automation Orchestrator - unified entry point for all automation operations.
 * Integrates all automation components: engine, adapters, memory, recovery, monitoring.
 */
public class AutomationOrchestrator {

    private static final String TAG = "AutomationOrchestrator";

    private final Context context;

    // Core components
    private HybridAutomationEngine engine;
    private AppAdapterManager adapterManager;
    private TemplateRegistry templateRegistry;

    // Integration components
    private MemoryAwareAutomation memoryAutomation;
    private IntentTriggeredAutomation intentAutomation;
    private SkillAutomationBridge skillBridge;

    // Support components
    private ErrorRecovery errorRecovery;
    private RetryPolicy defaultRetryPolicy;
    private PerformanceMonitor performanceMonitor;
    private AutomationLogger logger;

    // Configuration
    private AutomationConfig config;
    private boolean initialized = false;

    /**
     * Create orchestrator
     */
    public AutomationOrchestrator(@NonNull Context context) {
        this.context = context;
    }

    /**
     * Initialize all components
     */
    public void initialize(@NonNull AutomationConfig config,
                           @Nullable UserMemoryManager memoryManager,
                           @Nullable IntentEngine intentEngine) {
        Log.i(TAG, "Initializing Automation Orchestrator...");

        this.config = config;

        // Initialize core engine
        engine = new HybridAutomationEngine(context);
        engine.initialize(config);

        // Initialize adapter manager
        adapterManager = new AppAdapterManager();

        // Initialize template registry
        templateRegistry = new TemplateRegistry(adapterManager);

        // Initialize integration components
        if (memoryManager != null) {
            memoryAutomation = new MemoryAwareAutomation(context, engine, memoryManager);
        }

        if (intentEngine != null) {
            intentAutomation = new IntentTriggeredAutomation(context, engine, intentEngine);
        }

        skillBridge = new SkillAutomationBridge(
            context, engine, adapterManager, templateRegistry, memoryManager);

        // Initialize support components
        errorRecovery = new ErrorRecovery(engine);
        defaultRetryPolicy = RetryPolicy.standard();
        performanceMonitor = new PerformanceMonitor();
        logger = new AutomationLogger(context);

        initialized = true;
        Log.i(TAG, "Automation Orchestrator initialized");

        // Log capability
        Log.i(TAG, engine.getStatusSummary());
    }

    /**
     * Initialize with defaults
     */
    public void initialize() {
        initialize(AutomationConfig.builder().build(), null, null);
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        if (!initialized) return;

        Log.i(TAG, "Shutting down Automation Orchestrator...");

        if (engine != null) engine.shutdown();
        if (intentAutomation != null) intentAutomation.shutdown();
        if (logger != null) logger.shutdown();

        initialized = false;
    }

    // ===== Core Operations =====

    /**
     * Execute operation with retry and recovery
     */
    @NonNull
    public AutomationResult execute(@NonNull String operation,
                                     @NonNull Map<String, String> params) {
        if (!initialized) {
            return new AutomationResult(operation, "Orchestrator not initialized");
        }

        PerformanceMonitor.OperationTimer timer = performanceMonitor.startOperation(operation);

        try {
            AutomationResult result = executeWithRetry(operation, params);
            timer.complete(result.isSuccess(), result.getMessage(), params);

            // Log operation
            logger.info(operation, result.isSuccess() ? "Success" : "Failed: " + result.getMessage());

            return result;
        } catch (Exception e) {
            timer.failure(e.getMessage());
            logger.error(operation, "Exception", e.getMessage());
            return new AutomationResult(operation, e.getMessage());
        }
    }

    /**
     * Execute with retry logic
     */
    @NonNull
    private AutomationResult executeWithRetry(@NonNull String operation,
                                               @NonNull Map<String, String> params) {
        RetryPolicy policy = defaultRetryPolicy;
        AutomationResult result;

        do {
            // Get current adapter
            AppAdapter adapter = adapterManager.detectAdapter(engine);

            // Execute via adapter if available
            if (adapter != null && adapter.getSupportedOperations().contains(operation)) {
                result = adapterManager.execute(engine, operation, params);
            } else {
                // Execute via template
                result = templateRegistry.execute(engine, operation, params);
            }

            if (result.isSuccess()) {
                return result;
            }

            // Try recovery
            if (errorRecovery.isRecoverable(result.getMessage())) {
                AutomationResult recoveryResult = errorRecover(result);
                if (recoveryResult.isSuccess()) {
                    continue; // Retry after recovery
                }
            }

            // Wait before retry
            if (policy.shouldRetry(result)) {
                policy.waitBeforeRetry();
                policy.recordRetry();
            } else {
                break;
            }

        } while (true);

        return result;
    }

    /**
     * Attempt error recovery
     */
    @NonNull
    private AutomationResult errorRecover(@NonNull AutomationResult failedResult) {
        logger.warn("recovery", "Attempting recovery for: " + failedResult.getMessage());
        return errorRecovery.recover(failedResult);
    }

    // ===== Intent-based Execution =====

    /**
     * Process natural language and execute
     */
    @NonNull
    public AutomationResult processIntent(@NonNull String input) {
        if (intentAutomation == null) {
            return new AutomationResult("intent", "Intent engine not configured");
        }

        logger.info("intent", "Processing: " + input);
        return intentAutomation.processAndExecute(input);
    }

    // ===== Skill Execution =====

    /**
     * Execute a skill
     */
    @NonNull
    public AutomationResult executeSkill(@NonNull SkillDefinition skill,
                                          @NonNull Map<String, String> params) {
        if (!initialized) {
            return new AutomationResult(skill.getId(), "Orchestrator not initialized");
        }

        logger.info("skill", "Executing: " + skill.getId());
        return skillBridge.executeSkill(skill, params);
    }

    // ===== Template Execution =====

    /**
     * Execute template
     */
    @NonNull
    public AutomationResult executeTemplate(@NonNull String templateId,
                                             @NonNull Map<String, String> params) {
        if (!initialized) {
            return new AutomationResult(templateId, "Orchestrator not initialized");
        }

        logger.info("template", "Executing: " + templateId);
        return templateRegistry.execute(engine, templateId, params);
    }

    // ===== Memory-aware Operations =====

    /**
     * Get preferred shop for category
     */
    @Nullable
    public String getPreferredShop(@NonNull String category) {
        if (memoryAutomation == null) return null;
        return memoryAutomation.getPreferredShop(category);
    }

    /**
     * Remember preferred shop
     */
    public void rememberShop(@NonNull String category, @NonNull String shopName) {
        if (memoryAutomation != null) {
            memoryAutomation.rememberPreferredShop(category, shopName);
        }
    }

    // ===== Configuration =====

    /**
     * Set retry policy
     */
    public void setRetryPolicy(@NonNull RetryPolicy policy) {
        this.defaultRetryPolicy = policy;
    }

    /**
     * Set error recovery listener
     */
    public void setErrorRecoveryListener(@Nullable ErrorRecovery.RecoveryListener listener) {
        if (errorRecovery != null) {
            errorRecovery.setListener(listener);
        }
    }

    /**
     * Set performance listener
     */
    public void setPerformanceListener(@Nullable PerformanceMonitor.PerformanceListener listener) {
        if (performanceMonitor != null) {
            performanceMonitor.setListener(listener);
        }
    }

    /**
     * Set log callback
     */
    public void setLogCallback(@Nullable AutomationLogger.LogCallback callback) {
        if (logger != null) {
            logger.setCallback(callback);
        }
    }

    // ===== Status & Reporting =====

    /**
     * Check if initialized
     */
    public boolean isInitialized() {
        return initialized;
    }

    /**
     * Check if available
     */
    public boolean isAvailable() {
        return initialized && engine != null && engine.isAvailable();
    }

    /**
     * Get capability
     */
    @NonNull
    public AutomationCapability getCapability() {
        return engine != null ? engine.getCapability() : AutomationCapability.NONE;
    }

    /**
     * Get current adapter
     */
    @Nullable
    public AppAdapter getCurrentAdapter() {
        return adapterManager != null ? adapterManager.getCurrentAdapter() : null;
    }

    /**
     * Get current page
     */
    @NonNull
    public String getCurrentPage() {
        AppAdapter adapter = getCurrentAdapter();
        if (adapter != null && engine != null) {
            return adapter.detectPage(engine);
        }
        return "unknown";
    }

    /**
     * Get performance stats
     */
    @NonNull
    public String getPerformanceReport() {
        return performanceMonitor != null ? performanceMonitor.generateReport() : "No data";
    }

    /**
     * Get logs
     */
    @NonNull
    public org.json.JSONArray getLogs() {
        return logger != null ? logger.exportLogs() : new org.json.JSONArray();
    }

    /**
     * Get full status report
     */
    @NonNull
    public String getStatusReport() {
        StringBuilder sb = new StringBuilder();
        sb.append("=== Automation Orchestrator Status ===\n\n");

        sb.append("Initialized: ").append(initialized).append("\n");
        sb.append("Available: ").append(isAvailable()).append("\n");
        sb.append("Capability: ").append(getCapability()).append("\n");

        if (engine != null) {
            sb.append("\n").append(engine.getStatusSummary());
        }

        if (adapterManager != null) {
            sb.append("\nAdapters: ").append(adapterManager.getAdapterCount()).append(" registered\n");
        }

        if (performanceMonitor != null) {
            sb.append("\nPerformance:\n");
            sb.append(performanceMonitor.generateReport());
        }

        return sb.toString();
    }

    // ===== Getters =====

    @Nullable
    public HybridAutomationEngine getEngine() {
        return engine;
    }

    @Nullable
    public AppAdapterManager getAdapterManager() {
        return adapterManager;
    }

    @Nullable
    public TemplateRegistry getTemplateRegistry() {
        return templateRegistry;
    }

    @Nullable
    public ErrorRecovery getErrorRecovery() {
        return errorRecovery;
    }

    @Nullable
    public PerformanceMonitor getPerformanceMonitor() {
        return performanceMonitor;
    }

    @Nullable
    public AutomationLogger getLogger() {
        return logger;
    }
}