package com.ofa.agent.tool.builtin.system;

import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.content.Context;
import android.content.Intent;
import android.os.Build;
import android.util.Log;

import androidx.core.app.NotificationCompat;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.Map;
import java.util.UUID;

/**
 * Notification Tool - send system notifications.
 */
public class NotificationTool implements ToolExecutor {

    private static final String TAG = "NotificationTool";
    private static final String CHANNEL_ID = "ofa_agent_channel";
    private static final String CHANNEL_NAME = "OFA Agent Notifications";

    private final Context context;
    private final NotificationManager notificationManager;

    public NotificationTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.notificationManager = (NotificationManager) context.getSystemService(Context.NOTIFICATION_SERVICE);

        createNotificationChannel();
    }

    @NonNull
    @Override
    public String getToolId() {
        return "notification";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Send system notifications";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "send");

        switch (operation.toLowerCase()) {
            case "send":
                return executeSend(args, ctx);
            case "cancel":
                return executeCancel(args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return notificationManager != null;
    }

    @Override
    public boolean requiresAuth() {
        return false;
    }

    @Override
    public boolean supportsOffline() {
        return true;
    }

    @Nullable
    @Override
    public String[] getRequiredPermissions() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            return new String[]{"android.permission.POST_NOTIFICATIONS"};
        }
        return null;
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        if (operation == null) return false;

        if ("send".equalsIgnoreCase(operation)) {
            return args.containsKey("title") && args.containsKey("message");
        }

        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 50;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getSendDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'send'");
            operation.put("default", "send");
            props.put("operation", operation);

            JSONObject title = new JSONObject();
            title.put("type", "string");
            title.put("description", "Notification title");
            props.put("title", title);

            JSONObject message = new JSONObject();
            message.put("type", "string");
            message.put("description", "Notification message body");
            props.put("message", message);

            JSONObject priority = new JSONObject();
            priority.put("type", "string");
            priority.put("description", "Priority: high, default, low");
            priority.put("default", "default");
            props.put("priority", priority);

            JSONObject notificationId = new JSONObject();
            notificationId.put("type", "integer");
            notificationId.put("description", "Optional notification ID for updates");
            props.put("notificationId", notificationId);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"title", "message"});
        return new ToolDefinition("notification.send", "Send a notification",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getCancelDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'cancel'");
            props.put("operation", operation);

            JSONObject notificationId = new JSONObject();
            notificationId.put("type", "integer");
            notificationId.put("description", "Notification ID to cancel");
            props.put("notificationId", notificationId);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"notificationId"});
        return new ToolDefinition("notification.cancel", "Cancel a notification",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeSend(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String title = getStringArg(args, "title", "OFA Agent");
        String message = getStringArg(args, "message", "");
        String priority = getStringArg(args, "priority", "default");
        int notificationId = getIntArg(args, "notificationId", generateNotificationId());

        if (message.isEmpty()) {
            return new ToolResult(getToolId(), "Missing message parameter");
        }

        try {
            // Build notification
            NotificationCompat.Builder builder = new NotificationCompat.Builder(context, CHANNEL_ID)
                    .setSmallIcon(android.R.drawable.ic_dialog_info)
                    .setContentTitle(title)
                    .setContentText(message)
                    .setStyle(new NotificationCompat.BigTextStyle().bigText(message))
                    .setAutoCancel(true);

            // Set priority
            switch (priority.toLowerCase()) {
                case "high":
                    builder.setPriority(NotificationCompat.PRIORITY_HIGH);
                    break;
                case "low":
                    builder.setPriority(NotificationCompat.PRIORITY_LOW);
                    break;
                default:
                    builder.setPriority(NotificationCompat.PRIORITY_DEFAULT);
            }

            // Add tap intent (opens app)
            Intent tapIntent = context.getPackageManager().getLaunchIntentForPackage(context.getPackageName());
            if (tapIntent != null) {
                tapIntent.setFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_CLEAR_TOP);
                PendingIntent pendingIntent = PendingIntent.getActivity(context, 0, tapIntent,
                        PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE);
                builder.setContentIntent(pendingIntent);
            }

            // Send notification
            notificationManager.notify(notificationId, builder.build());

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("notificationId", notificationId);
            output.put("title", title);
            output.put("message", message);
            output.put("priority", priority);
            output.put("sent", true);

            return new ToolResult(getToolId(), output, 20);

        } catch (Exception e) {
            Log.e(TAG, "Send notification failed", e);
            return new ToolResult(getToolId(), "Send failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeCancel(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        int notificationId = getIntArg(args, "notificationId", -1);

        if (notificationId < 0) {
            return new ToolResult(getToolId(), "Missing notificationId parameter");
        }

        try {
            notificationManager.cancel(notificationId);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("notificationId", notificationId);
            output.put("cancelled", true);

            return new ToolResult(getToolId(), output, 10);

        } catch (Exception e) {
            Log.e(TAG, "Cancel notification failed", e);
            return new ToolResult(getToolId(), "Cancel failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    private void createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel channel = new NotificationChannel(
                    CHANNEL_ID,
                    CHANNEL_NAME,
                    NotificationManager.IMPORTANCE_DEFAULT
            );
            channel.setDescription("Notifications from OFA Agent");

            notificationManager.createNotificationChannel(channel);
        }
    }

    private int generateNotificationId() {
        return (int) (System.currentTimeMillis() % 10000);
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }

    private int getIntArg(@NonNull Map<String, Object> args, @NonNull String key, int defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        if (value instanceof Number) return ((Number) value).intValue();
        try {
            return Integer.parseInt(value.toString());
        } catch (Exception e) {
            return defaultVal;
        }
    }
}