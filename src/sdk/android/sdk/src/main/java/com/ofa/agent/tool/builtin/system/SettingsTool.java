package com.ofa.agent.tool.builtin.system;

import android.content.Context;
import android.media.AudioManager;
import android.net.wifi.WifiManager;
import android.os.Build;
import android.provider.Settings;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.constraint.ConstraintType;
import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.Map;

/**
 * Settings Tool - read and modify device settings.
 * Supports: get, set operations for various settings.
 */
public class SettingsTool implements ToolExecutor {

    private static final String TAG = "SettingsTool";

    private final Context context;
    private final AudioManager audioManager;

    public SettingsTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.audioManager = (AudioManager) context.getSystemService(Context.AUDIO_SERVICE);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "settings";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Read and modify device settings";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "get");
        String setting = getStringArg(args, "setting", null);

        if (setting == null) {
            return new ToolResult(getToolId(), "Missing setting parameter");
        }

        switch (operation.toLowerCase()) {
            case "get":
                return executeGet(setting, ctx);
            case "set":
                return executeSet(setting, args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return true;
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
        return null; // Most settings readable without permission
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        String setting = getStringArg(args, "setting", null);

        if (operation == null || setting == null) return false;

        if ("set".equalsIgnoreCase(operation)) {
            return args.containsKey("value");
        }

        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 100;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getGetDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'get'");
            operation.put("default", "get");
            props.put("operation", operation);

            JSONObject setting = new JSONObject();
            setting.put("type", "string");
            setting.put("description", "Setting name: volume, brightness, wifi, bluetooth, airplane");
            props.put("setting", setting);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"setting"});
        return new ToolDefinition("settings.get", "Get device setting value",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getSetDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'set'");
            operation.put("default", "set");
            props.put("operation", operation);

            JSONObject setting = new JSONObject();
            setting.put("type", "string");
            setting.put("description", "Setting name: volume, brightness");
            props.put("setting", setting);

            JSONObject value = new JSONObject();
            value.put("type", "integer");
            value.put("description", "Setting value");
            props.put("value", value);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"setting", "value"});

        return new ToolDefinition("settings.set", "Set device setting value",
                schema, true, new ConstraintType[]{ConstraintType.DEVICE});
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeGet(@NonNull String setting, @NonNull ExecutionContext ctx) {
        try {
            Object value = null;

            switch (setting.toLowerCase()) {
                case "volume":
                    value = audioManager.getStreamVolume(AudioManager.STREAM_RING);
                    break;

                case "volume_max":
                    value = audioManager.getStreamMaxVolume(AudioManager.STREAM_RING);
                    break;

                case "volume_music":
                    value = audioManager.getStreamVolume(AudioManager.STREAM_MUSIC);
                    break;

                case "brightness":
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
                        value = Settings.System.getInt(context.getContentResolver(),
                                Settings.System.SCREEN_BRIGHTNESS);
                    } else {
                        value = getLegacyBrightness();
                    }
                    break;

                case "wifi":
                    WifiManager wifiManager = (WifiManager) context.getApplicationContext()
                            .getSystemService(Context.WIFI_SERVICE);
                    if (wifiManager != null) {
                        value = wifiManager.isWifiEnabled();
                    }
                    break;

                case "bluetooth":
                    // Requires BLUETOOTH permission
                    value = "requires_permission";
                    break;

                case "airplane":
                    if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.JELLY_BEAN_MR1) {
                        value = Settings.Global.getInt(context.getContentResolver(),
                                Settings.Global.AIRPLANE_MODE_ON, 0) == 1;
                    }
                    break;

                case "screen_timeout":
                    value = Settings.System.getInt(context.getContentResolver(),
                            Settings.System.SCREEN_OFF_TIMEOUT);
                    break;

                default:
                    // Try to get from Settings.System
                    try {
                        value = Settings.System.getString(context.getContentResolver(), setting);
                    } catch (Exception e) {
                        value = null;
                    }
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("setting", setting);
            output.put("value", value);

            return new ToolResult(getToolId(), output, 50);

        } catch (Settings.SettingNotFoundException e) {
            return new ToolResult(getToolId(), "Setting not found: " + setting);
        } catch (Exception e) {
            Log.e(TAG, "Get setting failed", e);
            return new ToolResult(getToolId(), "Get failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeSet(@NonNull String setting, @NonNull Map<String, Object> args,
                                   @NonNull ExecutionContext ctx) {
        Object valueObj = args.get("value");
        if (valueObj == null) {
            return new ToolResult(getToolId(), "Missing value parameter");
        }

        try {
            boolean success = false;

            switch (setting.toLowerCase()) {
                case "volume":
                    int volume = (valueObj instanceof Number)
                            ? ((Number) valueObj).intValue()
                            : Integer.parseInt(valueObj.toString());

                    int maxVolume = audioManager.getStreamMaxVolume(AudioManager.STREAM_RING);
                    volume = Math.max(0, Math.min(volume, maxVolume));

                    audioManager.setStreamVolume(AudioManager.STREAM_RING, volume,
                            AudioManager.FLAG_SHOW_UI);
                    success = true;
                    break;

                case "volume_music":
                    int musicVolume = (valueObj instanceof Number)
                            ? ((Number) valueObj).intValue()
                            : Integer.parseInt(valueObj.toString());

                    int maxMusicVolume = audioManager.getStreamMaxVolume(AudioManager.STREAM_MUSIC);
                    musicVolume = Math.max(0, Math.min(musicVolume, maxMusicVolume));

                    audioManager.setStreamVolume(AudioManager.STREAM_MUSIC, musicVolume,
                            AudioManager.FLAG_SHOW_UI);
                    success = true;
                    break;

                case "brightness":
                    // Writing brightness requires WRITE_SETTINGS permission
                    int brightness = (valueObj instanceof Number)
                            ? ((Number) valueObj).intValue()
                            : Integer.parseInt(valueObj.toString());

                    brightness = Math.max(0, Math.min(brightness, 255));
                    success = Settings.System.putInt(context.getContentResolver(),
                            Settings.System.SCREEN_BRIGHTNESS, brightness);
                    break;

                case "screen_timeout":
                    int timeout = (valueObj instanceof Number)
                            ? ((Number) valueObj).intValue()
                            : Integer.parseInt(valueObj.toString());

                    success = Settings.System.putInt(context.getContentResolver(),
                            Settings.System.SCREEN_OFF_TIMEOUT, timeout);
                    break;

                default:
                    return new ToolResult(getToolId(), "Cannot set setting: " + setting);
            }

            JSONObject output = new JSONObject();
            output.put("success", success);
            output.put("setting", setting);
            output.put("value", valueObj);

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "Set setting failed", e);
            return new ToolResult(getToolId(), "Set failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    private int getLegacyBrightness() {
        try {
            return Settings.System.getInt(context.getContentResolver(),
                    Settings.System.SCREEN_BRIGHTNESS);
        } catch (Exception e) {
            return 128; // Default medium brightness
        }
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }
}