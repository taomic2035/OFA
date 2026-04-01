package com.ofa.agent.tool.builtin.device;

import android.content.Context;
import android.os.BatteryManager;
import android.os.Build;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.Map;

/**
 * Battery Tool - get battery status and information.
 */
public class BatteryTool implements ToolExecutor {

    private static final String TAG = "BatteryTool";

    private final Context context;
    private final BatteryManager batteryManager;

    public BatteryTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.batteryManager = (BatteryManager) context.getSystemService(Context.BATTERY_SERVICE);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "battery";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Get battery status and information";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        return executeStatus(ctx);
    }

    @Override
    public boolean isAvailable() {
        return batteryManager != null;
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
        return null;
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 20;
    }

    // ===== Tool Definition =====

    @NonNull
    public static ToolDefinition getStatusDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'status' (default)");
        return new ToolDefinition("battery.status", "Get battery status",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeStatus(@NonNull ExecutionContext ctx) {
        try {
            JSONObject output = new JSONObject();

            // Battery level
            int level = batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY);
            output.put("level", level);
            output.put("levelPercent", level + "%");

            // Charging status
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
                output.put("charging", batteryManager.isCharging());
            } else {
                // Legacy method - use Intent
                android.content.IntentFilter filter = new android.content.IntentFilter(android.content.Intent.ACTION_BATTERY_CHANGED);
                android.content.Intent batteryStatus = context.registerReceiver(null, filter);
                if (batteryStatus != null) {
                    int status = batteryStatus.getIntExtra(BatteryManager.EXTRA_STATUS, -1);
                    output.put("charging", status == BatteryManager.BATTERY_STATUS_CHARGING ||
                            status == BatteryManager.BATTERY_STATUS_FULL);
                }
            }

            // Battery health
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
                output.put("energyRemaining", batteryManager.getLongProperty(BatteryManager.BATTERY_PROPERTY_ENERGY_COUNTER));
            }

            // Additional info from Intent
            android.content.IntentFilter filter = new android.content.IntentFilter(android.content.Intent.ACTION_BATTERY_CHANGED);
            android.content.Intent batteryIntent = context.registerReceiver(null, filter);

            if (batteryIntent != null) {
                int status = batteryIntent.getIntExtra(BatteryManager.EXTRA_STATUS, -1);
                int plugged = batteryIntent.getIntExtra(BatteryManager.EXTRA_PLUGGED, -1);
                int health = batteryIntent.getIntExtra(BatteryManager.EXTRA_HEALTH, -1);
                int voltage = batteryIntent.getIntExtra(BatteryManager.EXTRA_VOLTAGE, -1);
                int temperature = batteryIntent.getIntExtra(BatteryManager.EXTRA_TEMPERATURE, -1);
                String technology = batteryIntent.getStringExtra(BatteryManager.EXTRA_TECHNOLOGY);

                output.put("status", getBatteryStatusName(status));
                output.put("plugged", getPluggedName(plugged));
                output.put("health", getBatteryHealthName(health));
                output.put("voltage", voltage);
                output.put("voltageVolts", voltage / 1000.0);
                output.put("temperature", temperature);
                output.put("temperatureCelsius", temperature / 10.0);
                output.put("technology", technology);
            }

            // Interpretations
            output.put("low", level < 20);
            output.put("critical", level < 10);
            output.put("full", level >= 100);

            output.put("success", true);

            return new ToolResult(getToolId(), output, 20);

        } catch (Exception e) {
            Log.e(TAG, "Get battery status failed", e);
            return new ToolResult(getToolId(), "Get status failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    @NonNull
    private String getBatteryStatusName(int status) {
        switch (status) {
            case BatteryManager.BATTERY_STATUS_UNKNOWN: return "unknown";
            case BatteryManager.BATTERY_STATUS_CHARGING: return "charging";
            case BatteryManager.BATTERY_STATUS_DISCHARGING: return "discharging";
            case BatteryManager.BATTERY_STATUS_NOT_CHARGING: return "not_charging";
            case BatteryManager.BATTERY_STATUS_FULL: return "full";
            default: return "unknown_" + status;
        }
    }

    @NonNull
    private String getPluggedName(int plugged) {
        switch (plugged) {
            case 0: return "unplugged";
            case BatteryManager.BATTERY_PLUGGED_AC: return "ac";
            case BatteryManager.BATTERY_PLUGGED_USB: return "usb";
            case BatteryManager.BATTERY_PLUGGED_WIRELESS: return "wireless";
            default: return "unknown_" + plugged;
        }
    }

    @NonNull
    private String getBatteryHealthName(int health) {
        switch (health) {
            case BatteryManager.BATTERY_HEALTH_UNKNOWN: return "unknown";
            case BatteryManager.BATTERY_HEALTH_GOOD: return "good";
            case BatteryManager.BATTERY_HEALTH_OVERHEAT: return "overheat";
            case BatteryManager.BATTERY_HEALTH_DEAD: return "dead";
            case BatteryManager.BATTERY_HEALTH_OVER_VOLTAGE: return "over_voltage";
            case BatteryManager.BATTERY_HEALTH_UNSPECIFIED_FAILURE: return "unspecified_failure";
            case BatteryManager.BATTERY_HEALTH_COLD: return "cold";
            default: return "unknown_" + health;
        }
    }
}