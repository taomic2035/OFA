package com.ofa.agent.example;

import android.app.Application;
import android.os.Bundle;
import android.util.Log;
import android.widget.Button;
import android.widget.TextView;
import android.widget.Toast;

import androidx.appcompat.app.AppCompatActivity;
import androidx.lifecycle.Observer;

import com.ofa.agent.core.OFAAndroidAgent;
import com.ofa.agent.core.AgentProfile;
import com.ofa.agent.core.TaskRequest;
import com.ofa.agent.core.TaskResult;
import com.ofa.agent.core.CenterConnection;
import com.ofa.agent.core.ErrorHandler;
import com.ofa.agent.core.ErrorHandler.OFAError;
import com.ofa.agent.identity.PersonalIdentity;
import com.ofa.agent.distributed.DistributedOrchestrator;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * OFA Android SDK Complete Usage Example.
 *
 * This example demonstrates:
 * 1. SDK initialization with various configurations
 * 2. Natural language task execution
 * 3. Skill-based execution
 * 4. Automation execution
 * 5. Identity management
 * 6. Behavior observation
 * 7. Peer communication
 * 8. Error handling and recovery
 * 9. Scene detection
 * 10. Distributed coordination
 *
 * @version 1.4.0
 */
public class OFAUsageExample extends AppCompatActivity {

    private static final String TAG = "OFAExample";

    private OFAAndroidAgent agent;
    private TextView statusTextView;
    private TextView resultTextView;

    // ===== 1. SDK Initialization =====

    /**
     * Example: Basic initialization (recommended for most apps)
     */
    public void initializeBasic() {
        agent = new OFAAndroidAgent.Builder(this)
            .runMode(AgentProfile.RunMode.SYNC) // SYNC mode: connected to Center
            .center("center.ofa.ai", 9090)
            .enableAutomation(true)
            .enableSocial(true)
            .enablePeerNetwork(true)
            .enableDistributed(true)
            .build();

        agent.initialize();
    }

    /**
     * Example: Standalone mode (offline-first, no Center dependency)
     */
    public void initializeStandalone() {
        agent = new OFAAndroidAgent.Builder(this)
            .runMode(AgentProfile.RunMode.STANDALONE) // No Center connection
            .enableAutomation(true)
            .enableSocial(false) // No social features in standalone
            .build();

        agent.initialize();
    }

    /**
     * Example: Hybrid mode (local-first with cloud enhancement)
     */
    public void initializeHybrid() {
        agent = new OFAAndroidAgent.Builder(this)
            .runMode(AgentProfile.RunMode.HYBRID) // Try local first, fallback to Center
            .center("center.ofa.ai", 9090)
            .enableAutomation(true)
            .enableDistributed(true)
            .build();

        agent.initialize();
    }

    /**
     * Example: Initialize in Application class (recommended)
     */
    public class OFAApplication extends Application {
        @Override
        public void onCreate() {
            super.onCreate();

            // Initialize agent early
            OFAAndroidAgent agent = new OFAAndroidAgent.Builder(this)
                .runMode(AgentProfile.RunMode.SYNC)
                .center("center.ofa.ai", 9090)
                .enableDistributed(true)
                .build();

            agent.initialize();

            // Add error listener
            ErrorHandler.addListener(new ErrorHandler.ErrorListener() {
                @Override
                public void onErrorOccurred(OFAError error) {
                    Log.w(TAG, "Error occurred: " + error.getMessage());
                }

                @Override
                public void onErrorRecovered(OFAError error) {
                    Log.i(TAG, "Error recovered: " + error.getMessage());
                }
            });
        }
    }

    // ===== 2. Natural Language Task Execution =====

    /**
     * Example: Execute natural language input
     */
    public void executeNaturalLanguage(String input) {
        CompletableFuture<TaskResult> future = agent.execute(input);

        future.whenComplete((result, error) -> {
            if (error == null) {
                showResult("Success: " + result.getData().get("output"));
                // Track success behavior
                agent.observeActivity("task_success", Map.of("input", input));
            } else {
                handleExecutionError(error, input);
            }
        });
    }

    /**
     * Example: Execute with timeout
     */
    public void executeWithTimeout(String input) {
        CompletableFuture<TaskResult> future = agent.execute(input);

        // Add timeout handling
        CompletableFuture<TaskResult> timeoutGuard = future
            .orTimeout(30, java.util.concurrent.TimeUnit.SECONDS);

        timeoutGuard.whenComplete((result, error) -> {
            if (error instanceof java.util.concurrent.TimeoutException) {
                showResult("Task timed out, using local fallback");
                // Use local execution
                executeLocalFallback(input);
            } else if (error != null) {
                handleExecutionError(error, input);
            } else {
                showResult("Success: " + result.getData().get("output"));
            }
        });
    }

    // ===== 3. Skill-Based Execution =====

    /**
     * Example: Execute a specific skill
     */
    public void executeSkill(String skillId, Map<String, String> inputs) {
        CompletableFuture<TaskResult> future = agent.executeSkill(skillId, inputs);

        future.whenComplete((result, error) -> {
            if (error == null) {
                showResult("Skill result: " + result.getData().get("output"));
            } else {
                handleExecutionError(error, skillId);
            }
        });
    }

    /**
     * Example: Search skill
     */
    public void searchForItem(String query) {
        Map<String, String> inputs = new HashMap<>();
        inputs.put("query", query);
        inputs.put("platform", "all");

        agent.executeSkill("search", inputs)
            .thenApply(result -> {
                Map<String, Object> data = result.getData();
                return data.get("results");
            })
            .thenAccept(results -> {
                showResult("Found " + results + " items");
            })
            .exceptionally(error -> {
                handleExecutionError(error, "search");
                return null;
            });
    }

    /**
     * Example: Order food skill
     */
    public void orderFood(String item, String restaurant) {
        Map<String, String> inputs = new HashMap<>();
        inputs.put("item", item);
        inputs.put("restaurant", restaurant);
        inputs.put("delivery_address", "home");

        agent.executeSkill("order_food", inputs)
            .thenAccept(result -> {
                if (result.getStatus() == TaskResult.Status.SUCCESS) {
                    showResult("Order placed: " + result.getData().get("order_id"));
                    // Record purchase behavior
                    agent.recordPurchase(item, 50.0, false);
                }
            });
    }

    // ===== 4. Automation Execution =====

    /**
     * Example: Execute automation operation
     */
    public void executeAutomation(String operation, Map<String, String> params) {
        agent.executeAutomation(operation, params)
            .thenAccept(result -> {
                showResult("Automation complete: " + result.getData().get("message"));
            });
    }

    /**
     * Example: Click text automation
     */
    public void clickText(String text) {
        Map<String, String> params = new HashMap<>();
        params.put("text", text);

        agent.executeAutomation("click_text", params)
            .thenAccept(result -> {
                showResult("Clicked '" + text + "'");
            });
    }

    /**
     * Example: Fill form automation
     */
    public void fillForm(Map<String, String> fields) {
        agent.executeAutomation("fill_form", fields)
            .thenAccept(result -> {
                showResult("Form filled successfully");
            });
    }

    // ===== 5. Identity Management =====

    /**
     * Example: Get user identity
     */
    public void getIdentity() {
        PersonalIdentity identity = agent.getIdentity();

        if (identity != null) {
            String info = "Identity: " + identity.getName() +
                         "\nID: " + identity.getId() +
                         "\nVersion: " + identity.getVersion();
            showResult(info);
        }
    }

    /**
     * Example: Get decision context for AI
     */
    public String getDecisionContext() {
        String context = agent.generatePromptContext();
        // Use this context when calling AI APIs
        return context;
    }

    /**
     * Example: Sync identity with Center
     */
    public void syncIdentity() {
        if (agent.isCenterConnected()) {
            agent.syncIdentity();
            showResult("Identity synced with Center");
        } else {
            showResult("Not connected to Center");
        }
    }

    // ===== 6. Behavior Observation =====

    /**
     * Example: Observe purchase decision
     */
    public void observePurchase(String item, double price, boolean isImpulse) {
        agent.observeDecision("purchase", Map.of(
            "item", item,
            "price", price,
            "is_impulse", isImpulse
        ));

        // Also use convenience method
        agent.recordPurchase(item, price, isImpulse);
    }

    /**
     * Example: Observe social interaction
     */
    public void observeSocialInteraction(String type, int participantCount, boolean usedEmoji) {
        agent.observeInteraction(type, Map.of(
            "participant_count", participantCount,
            "used_emoji", usedEmoji
        ));

        // Also use convenience method
        agent.recordSocialInteraction(type, participantCount, usedEmoji);
    }

    /**
     * Example: Observe preference change
     */
    public void observePreference(String preferenceType, Object oldValue, Object newValue) {
        agent.observePreference(preferenceType, Map.of(
            "old_value", oldValue,
            "new_value", newValue
        ));
    }

    /**
     * Example: Observe activity
     */
    public void observeActivity(String activityType, Map<String, Object> details) {
        agent.observeActivity(activityType, details);
    }

    // ===== 7. Peer Communication =====

    /**
     * Example: Get discovered peers
     */
    public void listPeers() {
        java.util.List<AgentProfile> peers = agent.getPeers();

        StringBuilder sb = new StringBuilder("Peers:\n");
        for (AgentProfile peer : peers) {
            sb.append("- ").append(peer.getName())
              .append(" (").append(peer.getType()).append(")\n");
        }
        showResult(sb.toString());
    }

    /**
     * Example: Send message to peer
     */
    public void sendToPeer(String peerId, String message) {
        boolean sent = agent.sendToPeer(peerId, message);
        if (sent) {
            showResult("Message sent to " + peerId);
        } else {
            showResult("Failed to send message");
        }
    }

    /**
     * Example: Request task from peer
     */
    public void requestFromPeer(String peerId, String skillId, Map<String, String> params) {
        TaskRequest request = TaskRequest.skill(skillId, params);
        TaskResult result = agent.requestFromPeer(peerId, request);

        if (result != null) {
            showResult("Peer result: " + result.getData().get("output"));
        }
    }

    // ===== 8. Error Handling =====

    /**
     * Example: Handle execution error with retry
     */
    public void handleExecutionError(Throwable error, String context) {
        OFAError ofaError = ErrorHandler.categorizeError(error);

        switch (ofaError.getStrategy()) {
            case IMMEDIATE_RETRY:
                // Retry immediately
                retryImmediately(context);
                break;

            case BACKOFF_RETRY:
                // Retry with backoff (handled by ErrorHandler)
                ErrorHandler.RetryExecutor retryExecutor =
                    new ErrorHandler.RetryExecutor(ErrorHandler.RetryConfig.defaultConfig());
                retryWithBackoff(retryExecutor, context);
                break;

            case GRACEFUL_DEGRADE:
                // Use fallback
                useFallback(ofaError, context);
                break;

            case CIRCUIT_BREAK:
                showResult("Service unavailable, please try later");
                break;

            case MANUAL_INTERVENTION:
                showResult("Error requires manual fix: " + ofaError.getMessage());
                break;

            case NONE:
                showResult("Unrecoverable error: " + ofaError.getMessage());
                break;
        }
    }

    /**
     * Example: Retry immediately
     */
    private void retryImmediately(String input) {
        agent.execute(input)
            .thenAccept(result -> showResult("Retry success: " + result.getData()));
    }

    /**
     * Example: Retry with exponential backoff
     */
    private void retryWithBackoff(ErrorHandler.RetryExecutor executor, String input) {
        executor.execute(
            attempt -> agent.execute(input),
            ErrorHandler.CircuitBreaker.defaultBreaker("task_execution")
        ).thenAccept(result -> showResult("Backoff retry success"));
    }

    /**
     * Example: Use fallback when main execution fails
     */
    private void useFallback(OFAError error, String context) {
        ErrorHandler.FallbackProvider fallback = ErrorHandler.FallbackProvider.createDefault();
        Object result = fallback.getFallback(error, context);

        if (result != null) {
            showResult("Fallback result: " + result);
        } else {
            executeLocalFallback(context);
        }
    }

    /**
     * Example: Execute locally when Center is unavailable
     */
    private void executeLocalFallback(String input) {
        // Use local execution engine
        TaskRequest request = TaskRequest.naturalLanguage(input);
        CompletableFuture<TaskResult> future = agent.getAutomationOrchestrator() != null ?
            agent.executeAutomation("process_local", Map.of("input", input)) :
            agent.execute(input);

        future.thenAccept(result -> showResult("Local execution: " + result.getData()));
    }

    /**
     * Example: Monitor circuit breaker state
     */
    public void monitorCircuitBreaker() {
        ErrorHandler.CircuitBreaker breaker =
            ErrorHandler.CircuitBreaker.defaultBreaker("center_connection");

        if (breaker.getState() == ErrorHandler.CircuitBreaker.State.OPEN) {
            showResult("Connection circuit is open, waiting...");
        } else if (breaker.getState() == ErrorHandler.CircuitBreaker.State.HALF_OPEN) {
            showResult("Connection recovering, testing...");
        } else {
            showResult("Connection circuit closed, normal operation");
        }
    }

    // ===== 9. Distributed Coordination =====

    /**
     * Example: Get distributed orchestrator
     */
    public void useDistributedOrchestrator() {
        DistributedOrchestrator orchestrator = agent.getDistributedOrchestrator();

        if (orchestrator != null) {
            // Get current scene
            String currentScene = orchestrator.getCurrentScene();
            showResult("Current scene: " + currentScene);

            // Get device roles
            Map<String, String> roles = orchestrator.getDeviceRoles();
            showResult("Device roles: " + roles);
        }
    }

    /**
     * Example: Handle scene change
     */
    public void onSceneChanged(String sceneType) {
        switch (sceneType) {
            case "running":
                // Route messages to phone
                // Filter watch notifications
                break;
            case "meeting":
                // Enable DND mode
                // Route reminders to glasses
                break;
            case "driving":
                // Enable car mode
                // Use voice commands only
                break;
        }
    }

    // ===== 10. Status Monitoring =====

    /**
     * Example: Get agent status report
     */
    public void getStatusReport() {
        String report = agent.getStatusReport();
        showResult(report);
    }

    /**
     * Example: Check connection status
     */
    public void checkConnection() {
        if (agent.isCenterConnected()) {
            showResult("Connected to Center");
        } else if (agent.isNetworkAvailable()) {
            showResult("Network available but Center not connected");
        } else {
            showResult("No network connection");
        }
    }

    /**
     * Example: Add status change listener
     */
    public void addStatusListener() {
        agent.addStatusChangeListener(new AgentModeManager.StatusChangeListener() {
            @Override
            public void onStatusChanged(AgentProfile.AgentStatus oldStatus,
                                         AgentProfile.AgentStatus newStatus) {
                Log.i(TAG, "Status changed: " + oldStatus + " -> " + newStatus);
                updateStatusDisplay(newStatus);
            }
        });

        agent.addModeChangeListener(new AgentModeManager.ModeChangeListener() {
            @Override
            public void onModeChanged(AgentProfile.RunMode oldMode,
                                      AgentProfile.RunMode newMode) {
                Log.i(TAG, "Mode changed: " + oldMode + " -> " + newMode);
                updateModeDisplay(newMode);
            }
        });
    }

    // ===== Helper Methods =====

    private void showResult(String message) {
        if (resultTextView != null) {
            resultTextView.setText(message);
        }
        Log.i(TAG, message);
    }

    private void updateStatusDisplay(AgentProfile.AgentStatus status) {
        if (statusTextView != null) {
            statusTextView.setText("Status: " + status.name());
        }
    }

    private void updateModeDisplay(AgentProfile.RunMode mode) {
        if (statusTextView != null) {
            statusTextView.setText("Mode: " + mode.name());
        }
    }

    // ===== Lifecycle =====

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Setup UI and initialize agent
        initializeBasic();
        addStatusListener();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (agent != null) {
            agent.shutdown();
        }
    }

    // ===== Best Practices =====

    /**
     * Best Practice 1: Always handle errors gracefully
     *
     * Don't:
     *   agent.execute(input).get(); // May block forever
     *
     * Do:
     *   agent.execute(input)
     *       .orTimeout(30, TimeUnit.SECONDS)
     *       .whenComplete((result, error) -> {
     *           if (error != null) handleExecutionError(error, input);
     *       });
     */

    /**
     * Best Practice 2: Use appropriate running mode
     *
     * - STANDALONE: Offline-only apps, no Center dependency
     * - SYNC: Always connected apps, Center-dependent
     * - HYBRID: Offline-first with cloud enhancement (recommended)
     */

    /**
     * Best Practice 3: Observe user behavior for personalization
     *
     * Always observe important decisions:
     * - Purchases (agent.recordPurchase)
     * - Social interactions (agent.recordSocialInteraction)
     * - Preferences changes (agent.observePreference)
     */

    /**
     * Best Practice 4: Initialize in Application class
     *
     * This ensures agent is ready before any activity starts.
     */

    /**
     * Best Practice 5: Shutdown properly
     *
     * Always call agent.shutdown() when app exits to:
     * - Disconnect from Center
     * - Save pending data
     * - Release resources
     */
}