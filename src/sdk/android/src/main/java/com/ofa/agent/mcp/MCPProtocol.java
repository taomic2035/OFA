package com.ofa.agent.mcp;

import androidx.annotation.NonNull;

import org.json.JSONObject;

/**
 * MCP Protocol constants and utility methods.
 * Implements Model Context Protocol specification.
 */
public final class MCPProtocol {

    // Protocol version
    public static final String VERSION = "2024-11-05";
    public static final String PROTOCOL_NAME = "mcp";

    // Message types
    public static final String MSG_REQUEST = "request";
    public static final String MSG_RESPONSE = "response";
    public static final String MSG_NOTIFICATION = "notification";

    // Method names - Tools
    public static final String METHOD_LIST_TOOLS = "tools/list";
    public static final String METHOD_CALL_TOOL = "tools/call";
    public static final String METHOD_GET_TOOL = "tools/get";

    // Method names - Resources
    public static final String METHOD_LIST_RESOURCES = "resources/list";
    public static final String METHOD_READ_RESOURCE = "resources/read";
    public static final String METHOD_GET_RESOURCE = "resources/get";

    // Method names - Prompts
    public static final String METHOD_LIST_PROMPTS = "prompts/list";
    public static final String METHOD_GET_PROMPT = "prompts/get";

    // Method names - Server
    public static final String METHOD_INITIALIZE = "initialize";
    public static final String METHOD_PING = "ping";
    public static final String METHOD_SHUTDOWN = "shutdown";

    // Error codes
    public static final int ERROR_INVALID_REQUEST = -32600;
    public static final int ERROR_METHOD_NOT_FOUND = -32601;
    public static final int ERROR_INVALID_PARAMS = -32602;
    public static final int ERROR_INTERNAL = -32603;
    public static final int ERROR_TOOL_NOT_FOUND = -32001;
    public static final int ERROR_TOOL_EXECUTION = -32002;
    public static final int ERROR_PERMISSION_DENIED = -32003;
    public static final int ERROR_CONSTRAINT_VIOLATION = -32004;
    public static final int ERROR_TIMEOUT = -32005;
    public static final int ERROR_OFFLINE_NOT_SUPPORTED = -32006;

    // Content types
    public static final String CONTENT_TEXT = "text";
    public static final String CONTENT_IMAGE = "image";
    public static final String CONTENT_RESOURCE = "resource";

    // Private constructor
    private MCPProtocol() {}

    /**
     * Create a request message
     */
    @NonNull
    public static JSONObject createRequest(@NonNull String method, @Nullable JSONObject params) {
        JSONObject request = new JSONObject();
        try {
            request.put("jsonrpc", "2.0");
            request.put("method", method);
            request.put("id", generateRequestId());
            if (params != null) {
                request.put("params", params);
            }
        } catch (Exception e) {
            // Should not fail
        }
        return request;
    }

    /**
     * Create a response message
     */
    @NonNull
    public static JSONObject createResponse(@NonNull String requestId, @Nullable JSONObject result) {
        JSONObject response = new JSONObject();
        try {
            response.put("jsonrpc", "2.0");
            response.put("id", requestId);
            if (result != null) {
                response.put("result", result);
            }
        } catch (Exception e) {
            // Should not fail
        }
        return response;
    }

    /**
     * Create an error response
     */
    @NonNull
    public static JSONObject createErrorResponse(@NonNull String requestId, int code,
                                                  @NonNull String message) {
        JSONObject response = new JSONObject();
        try {
            response.put("jsonrpc", "2.0");
            response.put("id", requestId);
            JSONObject error = new JSONObject();
            error.put("code", code);
            error.put("message", message);
            response.put("error", error);
        } catch (Exception e) {
            // Should not fail
        }
        return response;
    }

    /**
     * Create a notification message
     */
    @NonNull
    public static JSONObject createNotification(@NonNull String method, @Nullable JSONObject params) {
        JSONObject notification = new JSONObject();
        try {
            notification.put("jsonrpc", "2.0");
            notification.put("method", method);
            if (params != null) {
                notification.put("params", params);
            }
        } catch (Exception e) {
            // Should not fail
        }
        return notification;
    }

    /**
     * Generate unique request ID
     */
    @NonNull
    private static String generateRequestId() {
        return "req-" + System.currentTimeMillis() + "-" + (int)(Math.random() * 10000);
    }

    /**
     * Get error message for code
     */
    @NonNull
    public static String getErrorMessage(int code) {
        switch (code) {
            case ERROR_INVALID_REQUEST: return "Invalid Request";
            case ERROR_METHOD_NOT_FOUND: return "Method not found";
            case ERROR_INVALID_PARAMS: return "Invalid params";
            case ERROR_INTERNAL: return "Internal error";
            case ERROR_TOOL_NOT_FOUND: return "Tool not found";
            case ERROR_TOOL_EXECUTION: return "Tool execution failed";
            case ERROR_PERMISSION_DENIED: return "Permission denied";
            case ERROR_CONSTRAINT_VIOLATION: return "Constraint violation";
            case ERROR_TIMEOUT: return "Execution timeout";
            case ERROR_OFFLINE_NOT_SUPPORTED: return "Tool not available offline";
            default: return "Unknown error";
        }
    }

    /**
     * Build input schema for simple text parameter
     */
    @NonNull
    public static JSONObject buildSimpleTextSchema(@NonNull String paramName,
                                                    @NonNull String description) {
        JSONObject schema = new JSONObject();
        try {
            schema.put("type", "object");
            JSONObject props = new JSONObject();
            JSONObject param = new JSONObject();
            param.put("type", "string");
            param.put("description", description);
            props.put(paramName, param);
            schema.put("properties", props);
        } catch (Exception e) {
            // Should not fail
        }
        return schema;
    }

    /**
     * Build input schema with multiple parameters
     */
    @NonNull
    public static JSONObject buildObjectSchema(@NonNull JSONObject properties,
                                                @Nullable String[] required) {
        JSONObject schema = new JSONObject();
        try {
            schema.put("type", "object");
            schema.put("properties", properties);
            if (required != null && required.length > 0) {
                org.json.JSONArray reqArray = new org.json.JSONArray();
                for (String r : required) {
                    reqArray.put(r);
                }
                schema.put("required", reqArray);
            }
        } catch (Exception e) {
            // Should not fail
        }
        return schema;
    }
}