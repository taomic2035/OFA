package com.ofa.agent.tool;

import android.Manifest;
import android.app.Activity;
import android.content.Context;
import android.content.pm.PackageManager;
import android.os.Build;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.app.ActivityCompat;
import androidx.core.content.ContextCompat;

import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CopyOnWriteArrayList;

/**
 * Permission Manager - handles Android permission requests for tools.
 * Encapsulates permission checking and requesting logic.
 */
public class PermissionManager {

    private static final String TAG = "PermissionManager";

    private final Context context;
    private final Activity activity;
    private final List<PermissionRequest> pendingRequests = new CopyOnWriteArrayList<>();
    private final Map<String, Boolean> deniedPermissions = new HashMap<>();

    public PermissionManager(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.activity = null;
    }

    public PermissionManager(@NonNull Activity activity) {
        this.context = activity.getApplicationContext();
        this.activity = activity;
    }

    /**
     * Check if all permissions are granted
     * @param permissions Array of permission strings
     * @return true if all permissions are granted
     */
    public boolean checkPermissions(@NonNull String[] permissions) {
        for (String permission : permissions) {
            if (ContextCompat.checkSelfPermission(context, permission)
                    != PackageManager.PERMISSION_GRANTED) {
                return false;
            }
        }
        return true;
    }

    /**
     * Check single permission
     */
    public boolean checkPermission(@NonNull String permission) {
        return ContextCompat.checkSelfPermission(context, permission)
                == PackageManager.PERMISSION_GRANTED;
    }

    /**
     * Get missing permissions from a list
     */
    @NonNull
    public String[] getMissingPermissions(@NonNull String[] permissions) {
        List<String> missing = new ArrayList<>();
        for (String permission : permissions) {
            if (!checkPermission(permission)) {
                missing.add(permission);
            }
        }
        return missing.toArray(new String[0]);
    }

    /**
     * Request permissions (requires Activity)
     * @param permissions Permissions to request
     * @param requestCode Request code for callback
     * @param callback Result callback
     */
    public void requestPermissions(@NonNull String[] permissions, int requestCode,
                                    @Nullable PermissionCallback callback) {
        if (activity == null) {
            Log.w(TAG, "Cannot request permissions without Activity");
            if (callback != null) {
                callback.onDenied(permissions, new boolean[permissions.length]);
            }
            return;
        }

        String[] missing = getMissingPermissions(permissions);
        if (missing.length == 0) {
            if (callback != null) {
                callback.onGranted(permissions);
            }
            return;
        }

        // Store request for tracking
        PermissionRequest request = new PermissionRequest(requestCode, permissions, callback);
        pendingRequests.add(request);

        ActivityCompat.requestPermissions(activity, missing, requestCode);
    }

    /**
     * Request single permission
     */
    public void requestPermission(@NonNull String permission, int requestCode,
                                   @Nullable PermissionCallback callback) {
        requestPermissions(new String[]{permission}, requestCode, callback);
    }

    /**
     * Handle permission result from Activity
     * Call this from Activity.onRequestPermissionsResult
     */
    public void onRequestPermissionsResult(int requestCode, @NonNull String[] permissions,
                                            @NonNull int[] grantResults) {
        PermissionRequest request = findRequest(requestCode);
        if (request == null) {
            Log.w(TAG, "No pending request for code: " + requestCode);
            return;
        }

        pendingRequests.remove(request);

        boolean allGranted = true;
        boolean[] results = new boolean[permissions.length];

        for (int i = 0; i < permissions.length; i++) {
            results[i] = grantResults.length > i && grantResults[i] == PackageManager.PERMISSION_GRANTED;
            if (!results[i]) {
                allGranted = false;
                deniedPermissions.put(permissions[i], true);
            }
        }

        if (request.callback != null) {
            if (allGranted) {
                request.callback.onGranted(permissions);
            } else {
                request.callback.onDenied(permissions, results);
            }
        }
    }

    /**
     * Check if should show permission rationale
     */
    public boolean shouldShowRationale(@NonNull String permission) {
        if (activity == null) return false;
        return ActivityCompat.shouldShowRequestPermissionRationale(activity, permission);
    }

    /**
     * Check if permission was permanently denied
     */
    public boolean isPermanentlyDenied(@NonNull String permission) {
        return deniedPermissions.containsKey(permission) && !shouldShowRationale(permission);
    }

    /**
     * Clear denied permission record
     */
    public void clearDeniedRecord(@NonNull String permission) {
        deniedPermissions.remove(permission);
    }

    /**
     * Find pending request by code
     */
    @Nullable
    private PermissionRequest findRequest(int requestCode) {
        for (PermissionRequest request : pendingRequests) {
            if (request.requestCode == requestCode) {
                return request;
            }
        }
        return null;
    }

    /**
     * Get common permission groups
     */
    @NonNull
    public static String[] getCameraPermissions() {
        return new String[]{Manifest.permission.CAMERA};
    }

    @NonNull
    public static String[] getStoragePermissions() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
            // Android 11+ uses scoped storage
            return new String[]{}; // No permissions needed for app-private storage
        } else {
            return new String[]{
                Manifest.permission.READ_EXTERNAL_STORAGE,
                Manifest.permission.WRITE_EXTERNAL_STORAGE
            };
        }
    }

    @NonNull
    public static String[] getContactsPermissions() {
        return new String[]{
            Manifest.permission.READ_CONTACTS,
            Manifest.permission.WRITE_CONTACTS
        };
    }

    @NonNull
    public static String[] getCalendarPermissions() {
        return new String[]{
            Manifest.permission.READ_CALENDAR,
            Manifest.permission.WRITE_CALENDAR
        };
    }

    @NonNull
    public static String[] getLocationPermissions() {
        return new String[]{
            Manifest.permission.ACCESS_FINE_LOCATION,
            Manifest.permission.ACCESS_COARSE_LOCATION
        };
    }

    @NonNull
    public static String[] getBluetoothPermissions() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
            return new String[]{
                Manifest.permission.BLUETOOTH_SCAN,
                Manifest.permission.BLUETOOTH_CONNECT
            };
        } else {
            return new String[]{
                Manifest.permission.BLUETOOTH,
                Manifest.permission.BLUETOOTH_ADMIN
            };
        }
    }

    @NonNull
    public static String[] getPhonePermissions() {
        return new String[]{
            Manifest.permission.READ_PHONE_STATE
        };
    }

    @NonNull
    public static String[] getMicrophonePermissions() {
        return new String[]{
            Manifest.permission.RECORD_AUDIO
        };
    }

    @NonNull
    public static String[] getNFCPermissions() {
        return new String[]{}; // NFC doesn't require normal permissions
    }

    /**
     * Permission request tracking class
     */
    private static class PermissionRequest {
        final int requestCode;
        final String[] permissions;
        final PermissionCallback callback;

        PermissionRequest(int requestCode, String[] permissions, PermissionCallback callback) {
            this.requestCode = requestCode;
            this.permissions = permissions;
            this.callback = callback;
        }
    }

    /**
     * Permission callback interface
     */
    public interface PermissionCallback {
        void onGranted(@NonNull String[] permissions);
        void onDenied(@NonNull String[] permissions, @NonNull boolean[] grantResults);
    }
}