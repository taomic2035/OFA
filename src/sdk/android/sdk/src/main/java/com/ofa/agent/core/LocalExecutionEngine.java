package com.ofa.agent.core;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.automation.AutomationOrchestrator;
import com.ofa.agent.automation.AutomationResult;
import com.ofa.agent.intent.IntentEngine;
import com.ofa.agent.intent.UserIntent;
import com.ofa.agent.memory.UserMemoryManager;
import com.ofa.agent.skill.CompositeSkillExecutor;
import com.ofa.agent.skill.SkillDefinition;
import com.ofa.agent.skill.SkillRegistry;
import com.ofa.agent.social.SocialOrchestrator;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * Local Execution Engine - handles all local task execution.
 *
 * Capabilities:
 * - Intent recognition and execution
 * - Skill execution
 * - Automation workflows
 * - Social notifications
 * - Memory operations
 */
public class LocalExecutionEngine {

    private static final String TAG = "LocalExecutionEngine";

    private final Context context;
    private final UserMemoryManager memoryManager;
    private final AutomationOrchestrator automationOrchestrator;
    private final SocialOrchestrator socialOrchestrator;

    private IntentEngine intentEngine;
    private SkillRegistry skillRegistry;
    private CompositeSkillExecutor skillExecutor;

    private boolean initialized = false;

    /**
     * Create local execution engine
     */
    public LocalExecutionEngine(@NonNull Context context,
                                 @Nullable UserMemoryManager memoryManager,
                                 @Nullable AutomationOrchestrator automationOrchestrator,
                                 @Nullable SocialOrchestrator socialOrchestrator) {
        this.context = context;
        this.memoryManager = memoryManager;
        this.automationOrchestrator = automationOrchestrator;
        this.socialOrchestrator = socialOrchestrator;
    }

    /**
     * Initialize the engine
     */
    public void initialize() {
        Log.i(TAG, "Initializing local execution engine...");

        // Initialize intent engine
        intentEngine = new IntentEngine();

        // Initialize skill registry and executor
        skillRegistry = SkillRegistry.getInstance(context);
        skillExecutor = new CompositeSkillExecutor(context, null);

        initialized = true;
        Log.i(TAG, "Local execution engine initialized");
    }

    /**
     * Check if can execute a task locally
     */
    public boolean canExecute(@NonNull TaskRequest request) {
        String type = request.type;

        switch (type) {
            case TaskRequest.TYPE_INTENT:
            case TaskRequest.TYPE_SKILL:
            case TaskRequest.TYPE_AUTOMATION:
            case TaskRequest.TYPE_SOCIAL:
            case TaskRequest.TYPE_MEMORY:
                return true;

            case TaskRequest.TYPE_CLOUD_LLM:
                return false; // Requires cloud

            default:
                return false;
        }
    }

    /**
     * Execute a task locally
     */
    @NonNull
    public CompletableFuture<TaskResult> execute(@NonNull TaskRequest request) {
        Log.d(TAG, "Executing task: " + request.taskId + " type: " + request.type);

        if (!initialized) {
            return CompletableFuture.completedFuture(
                TaskResult.failure(request.taskId, "Engine not initialized"));
        }

        try {
            switch (request.type) {
                case TaskRequest.TYPE_INTENT:
                    return executeIntent(request);

                case TaskRequest.TYPE_SKILL:
                    return executeSkill(request);

                case TaskRequest.TYPE_AUTOMATION:
                    return executeAutomation(request);

                case TaskRequest.TYPE_SOCIAL:
                    return executeSocial(request);

                case TaskRequest.TYPE_MEMORY:
                    return executeMemory(request);

                case TaskRequest.TYPE_NATURAL_LANGUAGE:
                    return executeNaturalLanguage(request);

                default:
                    return CompletableFuture.completedFuture(
                        TaskResult.failure(request.taskId, "Unknown task type: " + request.type));
            }
        } catch (Exception e) {
            Log.e(TAG, "Task execution failed", e);
            return CompletableFuture.completedFuture(
                TaskResult.failure(request.taskId, e.getMessage()));
        }
    }

    /**
     * Execute intent-based task
     */
    @NonNull
    private CompletableFuture<TaskResult> executeIntent(@NonNull TaskRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String input = request.params.getOrDefault("input", "").toString();
                UserIntent intent = intentEngine.recognizeBest(input);

                if (intent == null) {
                    return TaskResult.failure(request.taskId, "Could not recognize intent");
                }

                Map<String, Object> result = new HashMap<>();
                result.put("intent", intent.getIntentName());
                result.put("slots", intent.getSlots());
                result.put("confidence", intent.getConfidence());

                return TaskResult.success(request.taskId, result);
            } catch (Exception e) {
                return TaskResult.failure(request.taskId, e.getMessage());
            }
        });
    }

    /**
     * Execute skill-based task
     */
    @NonNull
    private CompletableFuture<TaskResult> executeSkill(@NonNull TaskRequest request) {
        try {
            String skillId = request.params.getOrDefault("skillId", "").toString();
            SkillDefinition skill = skillRegistry.getSkill(skillId);

            if (skill == null) {
                return CompletableFuture.completedFuture(
                    TaskResult.failure(request.taskId, "Skill not found: " + skillId));
            }

            @SuppressWarnings("unchecked")
            Map<String, String> inputs = (Map<String, String>) request.params.getOrDefault("inputs", new HashMap<>());

            return skillExecutor.execute(skill, inputs)
                .thenApply(result -> {
                    if (result.isSuccess()) {
                        Map<String, Object> output = new HashMap<>();
                        output.put("output", result.getOutput());
                        return TaskResult.success(request.taskId, output);
                    } else {
                        return TaskResult.failure(request.taskId, result.getError());
                    }
                });
        } catch (Exception e) {
            return CompletableFuture.completedFuture(
                TaskResult.failure(request.taskId, e.getMessage()));
        }
    }

    /**
     * Execute automation task
     */
    @NonNull
    private CompletableFuture<TaskResult> executeAutomation(@NonNull TaskRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            if (automationOrchestrator == null) {
                return TaskResult.failure(request.taskId, "Automation not available");
            }

            try {
                String operation = request.params.getOrDefault("operation", "").toString();
                @SuppressWarnings("unchecked")
                Map<String, String> params = (Map<String, String>) request.params.getOrDefault("params", new HashMap<>());

                AutomationResult result;
                if (request.params.containsKey("templateId")) {
                    String templateId = request.params.get("templateId").toString();
                    result = automationOrchestrator.executeTemplate(templateId, params);
                } else {
                    result = automationOrchestrator.execute(operation, params);
                }

                if (result.isSuccess()) {
                    Map<String, Object> output = new HashMap<>();
                    output.put("result", result.getMessage());
                    return TaskResult.success(request.taskId, output);
                } else {
                    return TaskResult.failure(request.taskId, result.getMessage());
                }
            } catch (Exception e) {
                return TaskResult.failure(request.taskId, e.getMessage());
            }
        });
    }

    /**
     * Execute social notification task
     */
    @NonNull
    private CompletableFuture<TaskResult> executeSocial(@NonNull TaskRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            if (socialOrchestrator == null) {
                return TaskResult.failure(request.taskId, "Social notification not available");
            }

            try {
                String message = request.params.getOrDefault("message", "").toString();
                String recipient = request.params.getOrDefault("recipient", "").toString();
                String phone = request.params.getOrDefault("phone", "").toString();

                SocialOrchestrator.DeliveryRecord record =
                    socialOrchestrator.sendNotification(message, recipient, phone);

                Map<String, Object> output = new HashMap<>();
                output.put("success", record.success);
                output.put("channel", record.successfulChannel);
                output.put("type", record.messageType);

                if (record.success) {
                    return TaskResult.success(request.taskId, output);
                } else {
                    return TaskResult.failure(request.taskId, record.failureReason);
                }
            } catch (Exception e) {
                return TaskResult.failure(request.taskId, e.getMessage());
            }
        });
    }

    /**
     * Execute memory operation
     */
    @NonNull
    private CompletableFuture<TaskResult> executeMemory(@NonNull TaskRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            if (memoryManager == null) {
                return TaskResult.failure(request.taskId, "Memory system not available");
            }

            try {
                String operation = request.params.getOrDefault("operation", "get").toString();
                String key = request.params.getOrDefault("key", "").toString();

                Map<String, Object> output = new HashMap<>();

                switch (operation) {
                    case "get":
                        String value = memoryManager.get(key);
                        output.put("value", value != null ? value : "");
                        break;

                    case "set":
                        String newValue = request.params.getOrDefault("value", "").toString();
                        memoryManager.set(key, newValue);
                        output.put("success", true);
                        break;

                    case "delete":
                        memoryManager.delete(key);
                        output.put("success", true);
                        break;

                    case "suggestions":
                        int limit = (int) request.params.getOrDefault("limit", 5);
                        output.put("suggestions", memoryManager.getSuggestions(key, limit));
                        break;

                    default:
                        return TaskResult.failure(request.taskId, "Unknown memory operation: " + operation);
                }

                return TaskResult.success(request.taskId, output);
            } catch (Exception e) {
                return TaskResult.failure(request.taskId, e.getMessage());
            }
        });
    }

    /**
     * Execute natural language task (intent + execution)
     */
    @NonNull
    private CompletableFuture<TaskResult> executeNaturalLanguage(@NonNull TaskRequest request) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String input = request.params.getOrDefault("input", "").toString();

                // Step 1: Recognize intent
                UserIntent intent = intentEngine.recognizeBest(input);
                if (intent == null) {
                    return TaskResult.failure(request.taskId, "Could not understand: " + input);
                }

                // Step 2: Execute based on intent
                Map<String, Object> result = new HashMap<>();
                result.put("intent", intent.getIntentName());
                result.put("slots", intent.getSlots());

                // Try to execute via automation if available
                if (automationOrchestrator != null) {
                    String operation = mapIntentToOperation(intent.getIntentName());
                    if (operation != null) {
                        AutomationResult autoResult = automationOrchestrator.execute(
                            operation, intent.getSlots());
                        result.put("executed", autoResult.isSuccess());
                        result.put("message", autoResult.getMessage());
                    }
                }

                return TaskResult.success(request.taskId, result);
            } catch (Exception e) {
                return TaskResult.failure(request.taskId, e.getMessage());
            }
        });
    }

    /**
     * Map intent to automation operation
     */
    @Nullable
    private String mapIntentToOperation(@NonNull String intentName) {
        switch (intentName) {
            case "app_launch":
                return "launch";
            case "search":
                return "search";
            case "order_food":
                return "order_food";
            default:
                return null;
        }
    }

    /**
     * Process natural language input
     */
    @NonNull
    public CompletableFuture<TaskResult> processNaturalLanguage(@NonNull String input) {
        TaskRequest request = new TaskRequest.Builder()
            .type(TaskRequest.TYPE_NATURAL_LANGUAGE)
            .param("input", input)
            .build();

        return execute(request);
    }

    /**
     * Shutdown
     */
    public void shutdown() {
        Log.i(TAG, "Shutting down local execution engine...");
        initialized = false;
    }
}