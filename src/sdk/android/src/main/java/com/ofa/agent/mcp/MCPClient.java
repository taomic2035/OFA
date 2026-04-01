package com.ofa.agent.mcp;

import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;

/**
 * MCP Client - connects to external MCP servers.
 * Used for calling tools on remote MCP servers.
 */
public class MCPClient {

    private static final String TAG = "MCPClient";
    private static final int DEFAULT_TIMEOUT = 30000;

    private final String serverUrl;
    private final int timeoutMs;
    private final ExecutorService executor;
    private volatile boolean connected = false;

    public MCPClient(@NonNull String serverUrl) {
        this(serverUrl, DEFAULT_TIMEOUT);
    }

    public MCPClient(@NonNull String serverUrl, int timeoutMs) {
        this.serverUrl = serverUrl;
        this.timeoutMs = timeoutMs;
        this.executor = Executors.newCachedThreadPool();
    }

    /**
     * Initialize connection to MCP server
     */
    public void initialize(@Nullable InitializeCallback callback) {
        executor.execute(() -> {
            try {
                JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_INITIALIZE, null);
                JSONObject response = sendRequest(request);

                if (response.has("result")) {
                    connected = true;
                    if (callback != null) {
                        callback.onSuccess(response.getJSONObject("result"));
                    }
                    Log.i(TAG, "Connected to MCP server: " + serverUrl);
                } else {
                    connected = false;
                    String error = getErrorMessage(response);
                    if (callback != null) {
                        callback.onError(error);
                    }
                    Log.e(TAG, "Initialize failed: " + error);
                }
            } catch (Exception e) {
                connected = false;
                if (callback != null) {
                    callback.onError(e.getMessage());
                }
                Log.e(TAG, "Initialize error", e);
            }
        });
    }

    /**
     * List tools from server
     */
    @Nullable
    public List<ToolDefinition> listTools() {
        try {
            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_LIST_TOOLS, null);
            JSONObject response = sendRequest(request);

            if (response.has("result")) {
                JSONArray toolsArray = response.getJSONObject("result").getJSONArray("tools");
                List<ToolDefinition> tools = new java.util.ArrayList<>();
                for (int i = 0; i < toolsArray.length(); i++) {
                    ToolDefinition def = ToolDefinition.fromJson(toolsArray.getJSONObject(i));
                    if (def != null) {
                        tools.add(def);
                    }
                }
                return tools;
            }
        } catch (Exception e) {
            Log.e(TAG, "List tools error", e);
        }
        return null;
    }

    /**
     * Call a tool on the server
     */
    @Nullable
    public ToolResult callTool(@NonNull String toolName, @NonNull Map<String, Object> args) {
        try {
            JSONObject params = new JSONObject();
            params.put("name", toolName);
            params.put("arguments", mapToJson(args));

            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_CALL_TOOL, params);
            JSONObject response = sendRequest(request);

            if (response.has("result")) {
                JSONObject result = response.getJSONObject("result");
                return ToolResult.fromJson(result);
            } else {
                String error = getErrorMessage(response);
                return new ToolResult(toolName, error != null ? error : "Unknown error");
            }
        } catch (Exception e) {
            Log.e(TAG, "Call tool error", e);
            return new ToolResult(toolName, e.getMessage());
        }
    }

    /**
     * Call tool asynchronously
     */
    public void callToolAsync(@NonNull String toolName, @NonNull Map<String, Object> args,
                               @NonNull ToolCallback callback) {
        executor.execute(() -> {
            ToolResult result = callTool(toolName, args);
            if (result != null) {
                if (result.isSuccess()) {
                    callback.onSuccess(result);
                } else {
                    callback.onError(result.getError());
                }
            } else {
                callback.onError("Unknown error");
            }
        });
    }

    /**
     * List resources from server
     */
    @Nullable
    public List<ResourceDefinition> listResources() {
        try {
            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_LIST_RESOURCES, null);
            JSONObject response = sendRequest(request);

            if (response.has("result")) {
                JSONArray resourcesArray = response.getJSONObject("result").getJSONArray("resources");
                List<ResourceDefinition> resources = new java.util.ArrayList<>();
                for (int i = 0; i < resourcesArray.length(); i++) {
                    ResourceDefinition def = ResourceDefinition.fromJson(resourcesArray.getJSONObject(i));
                    if (def != null) {
                        resources.add(def);
                    }
                }
                return resources;
            }
        } catch (Exception e) {
            Log.e(TAG, "List resources error", e);
        }
        return null;
    }

    /**
     * Read resource from server
     */
    @Nullable
    public ResourceContent readResource(@NonNull String uri) {
        try {
            JSONObject params = new JSONObject();
            params.put("uri", uri);

            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_READ_RESOURCE, params);
            JSONObject response = sendRequest(request);

            if (response.has("result")) {
                JSONObject result = response.getJSONObject("result");
                String mimeType = result.optString("mimeType", "application/octet-stream");
                long lastModified = result.optLong("lastModified", System.currentTimeMillis());

                if (result.has("text")) {
                    return new ResourceContent(uri, mimeType, result.getString("text"), lastModified);
                } else if (result.has("data")) {
                    // Base64 encoded data
                    String dataStr = result.getString("data");
                    byte[] data = android.util.Base64.decode(dataStr, android.util.Base64.DEFAULT);
                    return new ResourceContent(uri, mimeType, data, lastModified);
                }
            }
        } catch (Exception e) {
            Log.e(TAG, "Read resource error", e);
        }
        return null;
    }

    /**
     * List prompts from server
     */
    @Nullable
    public List<PromptDefinition> listPrompts() {
        try {
            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_LIST_PROMPTS, null);
            JSONObject response = sendRequest(request);

            if (response.has("result")) {
                JSONArray promptsArray = response.getJSONObject("result").getJSONArray("prompts");
                List<PromptDefinition> prompts = new java.util.ArrayList<>();
                for (int i = 0; i < promptsArray.length(); i++) {
                    PromptDefinition def = PromptDefinition.fromJson(promptsArray.getJSONObject(i));
                    if (def != null) {
                        prompts.add(def);
                    }
                }
                return prompts;
            }
        } catch (Exception e) {
            Log.e(TAG, "List prompts error", e);
        }
        return null;
    }

    /**
     * Get prompt from server
     */
    @Nullable
    public PromptResult getPrompt(@NonNull String name, @Nullable Map<String, Object> args) {
        try {
            JSONObject params = new JSONObject();
            params.put("name", name);
            if (args != null) {
                params.put("arguments", mapToJson(args));
            }

            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_GET_PROMPT, params);
            JSONObject response = sendRequest(request);

            if (response.has("result")) {
                JSONObject result = response.getJSONObject("result");
                String description = result.optString("description", "");
                JSONArray messages = result.optJSONArray("messages");

                // Build rendered text from messages
                StringBuilder text = new StringBuilder();
                if (messages != null) {
                    for (int i = 0; i < messages.length(); i++) {
                        JSONObject msg = messages.getJSONObject(i);
                        JSONObject content = msg.getJSONObject("content");
                        if (content.optString("type").equals("text")) {
                            text.append(content.optString("text", ""));
                            if (i < messages.length() - 1) {
                                text.append("\n");
                            }
                        }
                    }
                }

                return new PromptResult(name, text.toString(), args);
            }
        } catch (Exception e) {
            Log.e(TAG, "Get prompt error", e);
        }
        return null;
    }

    /**
     * Ping the server
     */
    public boolean ping() {
        try {
            JSONObject request = MCPProtocol.createRequest(MCPProtocol.METHOD_PING, null);
            JSONObject response = sendRequest(request);
            return response.has("result");
        } catch (Exception e) {
            Log.e(TAG, "Ping error", e);
            return false;
        }
    }

    /**
     * Check if connected
     */
    public boolean isConnected() {
        return connected;
    }

    /**
     * Shutdown client
     */
    public void shutdown() {
        connected = false;
        executor.shutdown();
        Log.i(TAG, "MCP Client shutdown");
    }

    // ===== Internal Methods =====

    @NonNull
    private JSONObject sendRequest(@NonNull JSONObject request) throws Exception {
        URL url = new URL(serverUrl);
        HttpURLConnection conn = (HttpURLConnection) url.openConnection();

        conn.setRequestMethod("POST");
        conn.setRequestProperty("Content-Type", "application/json");
        conn.setConnectTimeout(timeoutMs);
        conn.setReadTimeout(timeoutMs);
        conn.setDoOutput(true);

        // Send request
        OutputStream os = conn.getOutputStream();
        os.write(request.toString().getBytes("UTF-8"));
        os.flush();
        os.close();

        // Read response
        int responseCode = conn.getResponseCode();
        BufferedReader reader;
        if (responseCode >= 200 && responseCode < 300) {
            reader = new BufferedReader(new InputStreamReader(conn.getInputStream()));
        } else {
            reader = new BufferedReader(new InputStreamReader(conn.getErrorStream()));
        }

        StringBuilder response = new StringBuilder();
        String line;
        while ((line = reader.readLine()) != null) {
            response.append(line);
        }
        reader.close();

        return new JSONObject(response.toString());
    }

    @Nullable
    private String getErrorMessage(@NonNull JSONObject response) {
        if (response.has("error")) {
            try {
                JSONObject error = response.getJSONObject("error");
                return error.optString("message", "Unknown error");
            } catch (Exception e) {
                return response.optString("error", "Unknown error");
            }
        }
        return null;
    }

    @NonNull
    private JSONObject mapToJson(@NonNull Map<String, Object> map) {
        JSONObject json = new JSONObject();
        try {
            for (Map.Entry<String, Object> entry : map.entrySet()) {
                if (entry.getValue() instanceof Map) {
                    json.put(entry.getKey(), mapToJson((Map<String, Object>) entry.getValue()));
                } else if (entry.getValue() instanceof List) {
                    JSONArray array = new JSONArray();
                    for (Object item : (List<?>) entry.getValue()) {
                        array.put(item);
                    }
                    json.put(entry.getKey(), array);
                } else {
                    json.put(entry.getKey(), entry.getValue());
                }
            }
        } catch (Exception e) {
            // Should not fail
        }
        return json;
    }

    // ===== Callbacks =====

    public interface InitializeCallback {
        void onSuccess(@NonNull JSONObject serverInfo);
        void onError(@NonNull String error);
    }

    public interface ToolCallback {
        void onSuccess(@NonNull ToolResult result);
        void onError(@NonNull String error);
    }
}