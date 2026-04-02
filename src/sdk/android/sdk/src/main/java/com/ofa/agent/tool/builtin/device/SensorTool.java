package com.ofa.agent.tool.builtin.device;

import android.content.Context;
import android.hardware.Sensor;
import android.hardware.SensorEvent;
import android.hardware.SensorEventListener;
import android.hardware.SensorManager;
import android.os.Build;
import android.os.Handler;
import android.os.Looper;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.List;
import java.util.Map;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;

/**
 * Sensor Tool - read device sensors.
 */
public class SensorTool implements ToolExecutor {

    private static final String TAG = "SensorTool";
    private static final int READ_TIMEOUT = 3000; // 3 seconds

    private final Context context;
    private final SensorManager sensorManager;

    public SensorTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.sensorManager = (SensorManager) context.getSystemService(Context.SENSOR_SERVICE);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "sensor";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Read device sensor values";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "list");

        switch (operation.toLowerCase()) {
            case "list":
                return executeList(ctx);
            case "read":
                return executeRead(args, ctx);
            case "info":
                return executeInfo(args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return sensorManager != null;
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
        return null; // Most sensors don't require permissions
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        if (operation == null) return false;

        if ("read".equalsIgnoreCase(operation) || "info".equalsIgnoreCase(operation)) {
            return args.containsKey("sensorType");
        }

        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 100;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'list'");
        return new ToolDefinition("sensor.list", "List available sensors",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getReadDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'read'");
            props.put("operation", operation);

            JSONObject sensorType = new JSONObject();
            sensorType.put("type", "string");
            sensorType.put("description", "Sensor type: accelerometer, gyroscope, light, proximity, etc.");
            props.put("sensorType", sensorType);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"sensorType"});
        return new ToolDefinition("sensor.read", "Read sensor values",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getInfoDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'info'");
            props.put("operation", operation);

            JSONObject sensorType = new JSONObject();
            sensorType.put("type", "string");
            sensorType.put("description", "Sensor type to get info");
            props.put("sensorType", sensorType);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"sensorType"});
        return new ToolDefinition("sensor.info", "Get sensor information",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeList(@NonNull ExecutionContext ctx) {
        List<Sensor> sensors = sensorManager.getSensorList(Sensor.TYPE_ALL);

        try {
            JSONArray sensorsArray = new JSONArray();
            for (Sensor sensor : sensors) {
                try {
                    JSONObject sensorJson = new JSONObject();
                    sensorJson.put("name", sensor.getName());
                    sensorJson.put("type", getSensorTypeName(sensor.getType()));
                    sensorJson.put("typeId", sensor.getType());
                    sensorJson.put("vendor", sensor.getVendor());
                    sensorJson.put("version", sensor.getVersion());
                    sensorJson.put("maxRange", sensor.getMaximumRange());
                    sensorJson.put("resolution", sensor.getResolution());
                    sensorJson.put("power", sensor.getPower());
                    sensorsArray.put(sensorJson);
                } catch (Exception e) {
                    // Skip this sensor on error
                }
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", sensors.size());
            output.put("sensors", sensorsArray);

            return new ToolResult(getToolId(), output, 50);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeRead(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String sensorTypeStr = getStringArg(args, "sensorType", "accelerometer");
        int sensorType = getSensorTypeFromString(sensorTypeStr);

        Sensor sensor = sensorManager.getDefaultSensor(sensorType);
        if (sensor == null) {
            return new ToolResult(getToolId(), "Sensor not available: " + sensorTypeStr);
        }

        try {
            // Use semaphore to wait for sensor data
            Semaphore semaphore = new Semaphore(0);
            float[] valuesHolder = new float[3];
            boolean[] receivedHolder = new boolean[1];

            SensorEventListener listener = new SensorEventListener() {
                @Override
                public void onSensorChanged(SensorEvent event) {
                    if (event.values != null) {
                        int len = Math.min(event.values.length, 3);
                        for (int i = 0; i < len; i++) {
                            valuesHolder[i] = event.values[i];
                        }
                        receivedHolder[0] = true;
                    }
                    semaphore.release();
                    sensorManager.unregisterListener(this);
                }

                @Override
                public void onAccuracyChanged(Sensor sensor, int accuracy) {
                    // Not needed
                }
            };

            sensorManager.registerListener(listener, sensor, SensorManager.SENSOR_DELAY_NORMAL);

            // Wait for data
            boolean received = semaphore.tryAcquire(READ_TIMEOUT, TimeUnit.MILLISECONDS);

            JSONObject output = new JSONObject();
            output.put("success", received);
            output.put("sensor", sensorTypeStr);
            output.put("sensorName", sensor.getName());

            if (received && receivedHolder[0]) {
                JSONArray valuesArray = new JSONArray();
                for (float v : valuesHolder) {
                    valuesArray.put(v);
                }
                output.put("values", valuesArray);

                // Add interpreted values for common sensors
                addInterpretedValues(output, sensorType, valuesHolder);
            } else {
                output.put("message", "No sensor data received within timeout");
            }

            return new ToolResult(getToolId(), output, READ_TIMEOUT);

        } catch (Exception e) {
            Log.e(TAG, "Read sensor failed", e);
            return new ToolResult(getToolId(), "Read failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeInfo(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String sensorTypeStr = getStringArg(args, "sensorType", "accelerometer");
        int sensorType = getSensorTypeFromString(sensorTypeStr);

        Sensor sensor = sensorManager.getDefaultSensor(sensorType);
        if (sensor == null) {
            return new ToolResult(getToolId(), "Sensor not available: " + sensorTypeStr);
        }

        try {
            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("name", sensor.getName());
            output.put("type", getSensorTypeName(sensor.getType()));
            output.put("typeId", sensor.getType());
            output.put("vendor", sensor.getVendor());
            output.put("version", sensor.getVersion());
            output.put("maxRange", sensor.getMaximumRange());
            output.put("resolution", sensor.getResolution());
            output.put("power", sensor.getPower());
            output.put("minDelay", sensor.getMinDelay());

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
                output.put("maxDelay", sensor.getMaxDelay());
                output.put("reportingMode", sensor.getReportingMode());
            }

            return new ToolResult(getToolId(), output, 20);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    private int getSensorTypeFromString(@NonNull String typeStr) {
        switch (typeStr.toLowerCase()) {
            case "accelerometer":
                return Sensor.TYPE_ACCELEROMETER;
            case "gyroscope":
                return Sensor.TYPE_GYROSCOPE;
            case "magnetic_field":
            case "magnetic":
            case "magnetometer":
                return Sensor.TYPE_MAGNETIC_FIELD;
            case "orientation":
                return Sensor.TYPE_ORIENTATION;
            case "light":
                return Sensor.TYPE_LIGHT;
            case "pressure":
            case "barometer":
                return Sensor.TYPE_PRESSURE;
            case "temperature":
                return Sensor.TYPE_AMBIENT_TEMPERATURE;
            case "proximity":
                return Sensor.TYPE_PROXIMITY;
            case "gravity":
                return Sensor.TYPE_GRAVITY;
            case "linear_acceleration":
                return Sensor.TYPE_LINEAR_ACCELERATION;
            case "rotation_vector":
                return Sensor.TYPE_ROTATION_VECTOR;
            case "step_counter":
                return Sensor.TYPE_STEP_COUNTER;
            case "step_detector":
                return Sensor.TYPE_STEP_DETECTOR;
            case "humidity":
                return Sensor.TYPE_RELATIVE_HUMIDITY;
            default:
                try {
                    return Integer.parseInt(typeStr);
                } catch (NumberFormatException e) {
                    return Sensor.TYPE_ACCELEROMETER;
                }
        }
    }

    @NonNull
    private String getSensorTypeName(int type) {
        switch (type) {
            case Sensor.TYPE_ACCELEROMETER: return "accelerometer";
            case Sensor.TYPE_GYROSCOPE: return "gyroscope";
            case Sensor.TYPE_MAGNETIC_FIELD: return "magnetic_field";
            case Sensor.TYPE_ORIENTATION: return "orientation";
            case Sensor.TYPE_LIGHT: return "light";
            case Sensor.TYPE_PRESSURE: return "pressure";
            case Sensor.TYPE_AMBIENT_TEMPERATURE: return "temperature";
            case Sensor.TYPE_PROXIMITY: return "proximity";
            case Sensor.TYPE_GRAVITY: return "gravity";
            case Sensor.TYPE_LINEAR_ACCELERATION: return "linear_acceleration";
            case Sensor.TYPE_ROTATION_VECTOR: return "rotation_vector";
            case Sensor.TYPE_STEP_COUNTER: return "step_counter";
            case Sensor.TYPE_STEP_DETECTOR: return "step_detector";
            case Sensor.TYPE_RELATIVE_HUMIDITY: return "humidity";
            default: return "unknown_" + type;
        }
    }

    private void addInterpretedValues(@NonNull JSONObject output, int type, @NonNull float[] values) {
        try {
            switch (type) {
                case Sensor.TYPE_ACCELEROMETER:
                    output.put("x_mps2", values[0]);
                    output.put("y_mps2", values[1]);
                    output.put("z_mps2", values[2]);
                    output.put("magnitude_mps2", Math.sqrt(
                            values[0]*values[0] + values[1]*values[1] + values[2]*values[2]));
                    break;

                case Sensor.TYPE_LIGHT:
                    output.put("lux", values[0]);
                    break;

                case Sensor.TYPE_PROXIMITY:
                    output.put("cm", values[0]);
                    output.put("near", values[0] < 5);
                    break;

                case Sensor.TYPE_PRESSURE:
                    output.put("hPa", values[0]);
                    output.put("mbar", values[0]);
                    break;

                case Sensor.TYPE_AMBIENT_TEMPERATURE:
                    output.put("celsius", values[0]);
                    break;

                case Sensor.TYPE_RELATIVE_HUMIDITY:
                    output.put("percent", values[0]);
                    break;

                case Sensor.TYPE_STEP_COUNTER:
                    output.put("steps", (long) values[0]);
                    break;
            }
        } catch (Exception e) {
            // Should not fail
        }
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }
}