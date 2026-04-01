package com.ofa.agent.tool.builtin.device;

import android.Manifest;
import android.content.Context;
import android.content.pm.PackageManager;
import android.graphics.ImageFormat;
import android.hardware.camera2.CameraAccessException;
import android.hardware.camera2.CameraCaptureSession;
import android.hardware.camera2.CameraCharacteristics;
import android.hardware.camera2.CameraDevice;
import android.hardware.camera2.CameraManager;
import android.hardware.camera2.CaptureRequest;
import android.hardware.camera2.TotalCaptureResult;
import android.media.Image;
import android.media.ImageReader;
import android.os.Handler;
import android.os.HandlerThread;
import android.util.Log;
import android.util.Size;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.content.ContextCompat;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.FileOutputStream;
import java.nio.ByteBuffer;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.Semaphore;
import java.util.concurrent.TimeUnit;

/**
 * Camera Tool - capture photos and scan codes.
 * Uses Camera2 API for advanced camera control.
 */
public class CameraTool implements ToolExecutor {

    private static final String TAG = "CameraTool";
    private static final int CAPTURE_TIMEOUT = 5000; // 5 seconds

    private final Context context;
    private final CameraManager cameraManager;
    private final Semaphore cameraLock = new Semaphore(1);

    public CameraTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.cameraManager = (CameraManager) context.getSystemService(Context.CAMERA_SERVICE);
    }

    @NonNull
    @Override
    public String getToolId() {
        return "camera";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Capture photos and scan codes";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "capture");

        switch (operation.toLowerCase()) {
            case "capture":
                return executeCapture(args, ctx);
            case "scan":
                return executeScan(args, ctx);
            case "list":
                return executeList(ctx);
            case "info":
                return executeInfo(args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return cameraManager != null && hasCameraPermission();
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
        return new String[]{Manifest.permission.CAMERA};
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        if (operation == null) return false;

        switch (operation.toLowerCase()) {
            case "capture":
            case "scan":
                return true;
            case "list":
                return true;
            case "info":
                return args.containsKey("cameraId");
            default:
                return false;
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 3000;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getCaptureDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'capture'");
            operation.put("default", "capture");
            props.put("operation", operation);

            JSONObject cameraId = new JSONObject();
            cameraId.put("type", "string");
            cameraId.put("description", "Camera ID (0 for back, 1 for front)");
            cameraId.put("default", "0");
            props.put("cameraId", cameraId);

            JSONObject savePath = new JSONObject();
            savePath.put("type", "string");
            savePath.put("description", "Path to save photo (optional)");
            props.put("savePath", savePath);

            JSONObject returnBase64 = new JSONObject();
            returnBase64.put("type", "boolean");
            returnBase64.put("description", "Return image as base64");
            returnBase64.put("default", false);
            props.put("returnBase64", returnBase64);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("camera.capture", "Capture a photo",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getScanDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'scan'");
            operation.put("default", "scan");
            props.put("operation", operation);

            JSONObject cameraId = new JSONObject();
            cameraId.put("type", "string");
            cameraId.put("description", "Camera ID (0 for back, 1 for front)");
            cameraId.put("default", "0");
            props.put("cameraId", cameraId);

            JSONObject scanType = new JSONObject();
            scanType.put("type", "string");
            scanType.put("description", "Type to scan: qr, barcode, all");
            scanType.put("default", "all");
            props.put("scanType", scanType);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("camera.scan", "Scan QR code or barcode",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'list'");
        return new ToolDefinition("camera.list", "List available cameras",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeCapture(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (!hasCameraPermission()) {
            return new ToolResult(getToolId(), "Camera permission not granted");
        }

        String cameraId = getStringArg(args, "cameraId", "0");
        String savePath = getStringArg(args, "savePath", null);
        boolean returnBase64 = getBooleanArg(args, "returnBase64", false);

        try {
            // Acquire camera lock
            if (!cameraLock.tryAcquire(CAPTURE_TIMEOUT, TimeUnit.MILLISECONDS)) {
                return new ToolResult(getToolId(), "Camera busy");
            }

            // Start background thread
            HandlerThread thread = new HandlerThread("CameraCapture");
            thread.start();
            Handler handler = new Handler(thread.getLooper());

            // Get camera characteristics
            String[] cameraIds = cameraManager.getCameraIdList();
            if (!Arrays.asList(cameraIds).contains(cameraId)) {
                cameraId = cameraIds[0];
            }

            CameraCharacteristics characteristics = cameraManager.getCameraCharacteristics(cameraId);
            Size[] jpegSizes = characteristics.get(CameraCharacteristics.SCALER_STREAM_CONFIGURATION_MAP)
                    .getOutputSizes(ImageFormat.JPEG);

            Size targetSize = jpegSizes.length > 0 ? jpegSizes[0] : new Size(1920, 1080);

            // Create image reader
            ImageReader imageReader = ImageReader.newInstance(targetSize.getWidth(), targetSize.getHeight(),
                    ImageFormat.JPEG, 1);

            // Open camera
            CameraDevice[] cameraDeviceHolder = new CameraDevice[1];
            byte[] imageDataHolder = new byte[1];
            Exception[] errorHolder = new Exception[1];

            cameraManager.openCamera(cameraId, new CameraDevice.StateCallback() {
                @Override
                public void onOpened(@NonNull CameraDevice camera) {
                    cameraDeviceHolder[0] = camera;

                    try {
                        // Create capture request
                        CaptureRequest.Builder requestBuilder = camera.createCaptureRequest(CameraDevice.TEMPLATE_STILL_CAPTURE);
                        requestBuilder.addTarget(imageReader.getSurface());

                        // Create session
                        camera.createCaptureSession(Arrays.asList(imageReader.getSurface()),
                                new CameraCaptureSession.StateCallback() {
                                    @Override
                                    public void onConfigured(@NonNull CameraCaptureSession session) {
                                        try {
                                            session.capture(requestBuilder.build(),
                                                    new CameraCaptureSession.CaptureCallback() {
                                                        @Override
                                                        public void onCaptureCompleted(@NonNull CameraCaptureSession session,
                                                                                        @NonNull CaptureRequest request,
                                                                                        @NonNull TotalCaptureResult result) {
                                                            // Image captured
                                                        }
                                                    }, handler);
                                        } catch (Exception e) {
                                            errorHolder[0] = e;
                                        }
                                    }

                                    @Override
                                    public void onConfigureFailed(@NonNull CameraCaptureSession session) {
                                        errorHolder[0] = new Exception("Session configuration failed");
                                    }
                                }, handler);
                    } catch (Exception e) {
                        errorHolder[0] = e;
                    }
                }

                @Override
                public void onDisconnected(@NonNull CameraDevice camera) {
                    camera.close();
                }

                @Override
                public void onError(@NonNull CameraDevice camera, int error) {
                    camera.close();
                    errorHolder[0] = new Exception("Camera error: " + error);
                }
            }, handler);

            // Wait for image
            Image image = imageReader.acquireLatestImage();
            if (image != null) {
                ByteBuffer buffer = image.getPlanes()[0].getBuffer();
                imageDataHolder[0] = new byte[buffer.remaining()];
                buffer.get(imageDataHolder[0]);
                image.close();
            }

            // Cleanup
            if (cameraDeviceHolder[0] != null) {
                cameraDeviceHolder[0].close();
            }
            imageReader.close();
            thread.quit();
            cameraLock.release();

            if (errorHolder[0] != null) {
                return new ToolResult(getToolId(), "Capture error: " + errorHolder[0].getMessage());
            }

            byte[] imageData = imageDataHolder[0];
            if (imageData == null || imageData.length == 0) {
                return new ToolResult(getToolId(), "No image captured");
            }

            // Save to file if path provided
            String savedPath = null;
            if (savePath != null) {
                File outputFile = resolvePath(savePath);
                FileOutputStream fos = new FileOutputStream(outputFile);
                fos.write(imageData);
                fos.close();
                savedPath = outputFile.getAbsolutePath();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("cameraId", cameraId);
            output.put("size", targetSize.getWidth() + "x" + targetSize.getHeight());
            output.put("bytes", imageData.length);

            if (savedPath != null) {
                output.put("savedPath", savedPath);
            }

            if (returnBase64) {
                output.put("imageBase64", android.util.Base64.encodeToString(imageData, android.util.Base64.DEFAULT));
            }

            return new ToolResult(getToolId(), output, 1500);

        } catch (Exception e) {
            cameraLock.release();
            Log.e(TAG, "Capture failed", e);
            return new ToolResult(getToolId(), "Capture failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeScan(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        // Note: Actual QR/barcode scanning requires additional library (e.g., ZXing)
        // This is a placeholder implementation

        JSONObject output = new JSONObject();
        output.put("success", false);
        output.put("message", "QR/barcode scanning requires additional library integration");

        return new ToolResult(getToolId(), output, 50);
    }

    @NonNull
    private ToolResult executeList(@NonNull ExecutionContext ctx) {
        try {
            String[] cameraIds = cameraManager.getCameraIdList();
            JSONArray camerasArray = new JSONArray();

            for (String id : cameraIds) {
                CameraCharacteristics characteristics = cameraManager.getCameraCharacteristics(id);
                Integer facing = characteristics.get(CameraCharacteristics.LENS_FACING);

                JSONObject camJson = new JSONObject();
                camJson.put("cameraId", id);
                camJson.put("facing", facing == CameraCharacteristics.LENS_FACING_BACK ? "back"
                        : facing == CameraCharacteristics.LENS_FACING_FRONT ? "front" : "external");

                // Get resolutions
                Size[] jpegSizes = characteristics.get(CameraCharacteristics.SCALER_STREAM_CONFIGURATION_MAP)
                        .getOutputSizes(ImageFormat.JPEG);
                JSONArray resolutions = new JSONArray();
                for (Size size : jpegSizes) {
                    resolutions.put(size.getWidth() + "x" + size.getHeight());
                }
                camJson.put("resolutions", resolutions);

                camerasArray.put(camJson);
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", cameraIds.length);
            output.put("cameras", camerasArray);

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "List cameras failed", e);
            return new ToolResult(getToolId(), "List failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeInfo(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String cameraId = getStringArg(args, "cameraId", "0");

        try {
            CameraCharacteristics characteristics = cameraManager.getCameraCharacteristics(cameraId);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("cameraId", cameraId);

            Integer facing = characteristics.get(CameraCharacteristics.LENS_FACING);
            output.put("facing", facing == CameraCharacteristics.LENS_FACING_BACK ? "back"
                    : facing == CameraCharacteristics.LENS_FACING_FRONT ? "front" : "external");

            Integer orientation = characteristics.get(CameraCharacteristics.SENSOR_ORIENTATION);
            output.put("orientation", orientation);

            Boolean flashAvailable = characteristics.get(CameraCharacteristics.FLASH_INFO_AVAILABLE);
            output.put("flashAvailable", flashAvailable);

            Float focalLength = characteristics.get(CameraCharacteristics.LENS_INFO_AVAILABLE_FOCAL_LENGTHS)[0];
            output.put("focalLength", focalLength);

            return new ToolResult(getToolId(), output, 50);

        } catch (Exception e) {
            Log.e(TAG, "Get camera info failed", e);
            return new ToolResult(getToolId(), "Get info failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    private boolean hasCameraPermission() {
        return ContextCompat.checkSelfPermission(context, Manifest.permission.CAMERA)
                == PackageManager.PERMISSION_GRANTED;
    }

    @NonNull
    private File resolvePath(@NonNull String path) {
        if (path.startsWith("/")) {
            return new File(path);
        }
        return new File(context.getExternalFilesDir(null), path);
    }

    @Nullable
    private String getStringArg(@NonNull Map<String, Object> args, @NonNull String key, @Nullable String defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        return value.toString();
    }

    private boolean getBooleanArg(@NonNull Map<String, Object> args, @NonNull String key, boolean defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        if (value instanceof Boolean) return (Boolean) value;
        return Boolean.parseBoolean(value.toString());
    }
}