package com.ofa.agent.automation.system;

import android.content.Context;
import android.content.Intent;
import android.net.Uri;
import android.os.Build;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.automation.AutomationResult;

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.io.IOException;

/**
 * Silent Installer - handles silent app installation and uninstallation.
 * Requires system-level permissions (INSTALL_PACKAGES, DELETE_PACKAGES) or root access.
 */
public class SilentInstaller {

    private static final String TAG = "SilentInstaller";

    private final Context context;
    private final SystemPermissionManager permissionManager;

    // Installation methods
    public static final int METHOD_PM_INSTALL = 1;      // Package manager (system app)
    public static final int METHOD_ROOT_INSTALL = 2;    // Root shell
    public static final int METHOD_SESSION_INSTALL = 3; // PackageInstaller session
    public static final int METHOD_USER_GUIDED = 4;     // Fallback to user interaction

    /**
     * Installation result with details
     */
    public static class InstallResult {
        public final boolean success;
        public final String packageName;
        public final int methodUsed;
        public final String message;
        public final long installTime;

        public InstallResult(boolean success, String packageName, int methodUsed, String message) {
            this.success = success;
            this.packageName = packageName;
            this.methodUsed = methodUsed;
            this.message = message;
            this.installTime = System.currentTimeMillis();
        }
    }

    /**
     * Create silent installer
     */
    public SilentInstaller(@NonNull Context context,
                           @NonNull SystemPermissionManager permissionManager) {
        this.context = context;
        this.permissionManager = permissionManager;
    }

    /**
     * Check if silent installation is possible
     */
    public boolean canInstallSilently() {
        return permissionManager.canSilentInstall();
    }

    /**
     * Install APK file
     */
    @NonNull
    public InstallResult install(@NonNull String apkPath) {
        Log.i(TAG, "Installing APK: " + apkPath);

        File apkFile = new File(apkPath);
        if (!apkFile.exists()) {
            return new InstallResult(false, null, METHOD_USER_GUIDED, "APK file not found");
        }

        // Determine best installation method
        int preferredMethod = determineInstallMethod();

        switch (preferredMethod) {
            case METHOD_PM_INSTALL:
                return installViaPackageManager(apkFile);

            case METHOD_ROOT_INSTALL:
                return installViaRoot(apkFile);

            case METHOD_SESSION_INSTALL:
                return installViaPackageInstaller(apkFile);

            default:
                return installViaUserGuided(apkFile);
        }
    }

    /**
     * Determine the best installation method
     */
    private int determineInstallMethod() {
        if (permissionManager.checkPermission(SystemPermissionManager.PERMISSION_INSTALL_PACKAGES)) {
            return METHOD_PM_INSTALL;
        }

        if (permissionManager.checkRootAccess()) {
            return METHOD_ROOT_INSTALL;
        }

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.LOLLIPOP) {
            return METHOD_SESSION_INSTALL;
        }

        return METHOD_USER_GUIDED;
    }

    /**
     * Install via package manager (requires system app)
     */
    @NonNull
    private InstallResult installViaPackageManager(@NonNull File apkFile) {
        Log.i(TAG, "Installing via Package Manager");

        try {
            // Copy APK to a temporary location that package manager can access
            File tempFile = copyToTempLocation(apkFile);

            // Use PackageManager.installPackage() (deprecated but works for system apps)
            // Modern approach would use PackageInstaller
            String command = "pm install -r " + tempFile.getAbsolutePath();

            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            // Cleanup temp file
            tempFile.delete();

            if (exitCode == 0) {
                String packageName = getPackageNameFromApk(apkFile);
                Log.i(TAG, "Installation successful: " + packageName);
                return new InstallResult(true, packageName, METHOD_PM_INSTALL, "Success");
            } else {
                return new InstallResult(false, null, METHOD_PM_INSTALL, "PM install failed");
            }
        } catch (Exception e) {
            Log.e(TAG, "PM install failed: " + e.getMessage());
            return new InstallResult(false, null, METHOD_PM_INSTALL, e.getMessage());
        }
    }

    /**
     * Install via root shell
     */
    @NonNull
    private InstallResult installViaRoot(@NonNull File apkFile) {
        Log.i(TAG, "Installing via Root");

        try {
            String command = "pm install -r " + apkFile.getAbsolutePath();

            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                String packageName = getPackageNameFromApk(apkFile);
                Log.i(TAG, "Root installation successful: " + packageName);
                return new InstallResult(true, packageName, METHOD_ROOT_INSTALL, "Success");
            } else {
                return new InstallResult(false, null, METHOD_ROOT_INSTALL, "Root install failed");
            }
        } catch (Exception e) {
            Log.e(TAG, "Root install failed: " + e.getMessage());
            return new InstallResult(false, null, METHOD_ROOT_INSTALL, e.getMessage());
        }
    }

    /**
     * Install via PackageInstaller session (Android 5.0+)
     */
    @NonNull
    private InstallResult installViaPackageInstaller(@NonNull File apkFile) {
        Log.i(TAG, "Installing via PackageInstaller session");

        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.LOLLIPOP) {
            return installViaUserGuided(apkFile);
        }

        try {
            android.content.pm.PackageInstaller packageInstaller =
                context.getPackageManager().getPackageInstaller();

            android.content.pm.PackageInstaller.SessionParams params =
                new android.content.pm.PackageInstaller.SessionParams(
                    android.content.pm.PackageInstaller.SessionParams.MODE_FULL_INSTALL);

            int sessionId = packageInstaller.createSession(params);
            android.content.pm.PackageInstaller.Session session = packageInstaller.openSession(sessionId);

            // Write APK to session
            FileOutputStream out = session.openWrite("apk", 0, apkFile.length());
            FileInputStream in = new FileInputStream(apkFile);
            byte[] buffer = new byte[4096];
            int bytesRead;
            while ((bytesRead = in.read(buffer)) != -1) {
                out.write(buffer, 0, bytesRead);
            }
            in.close();
            out.flush();
            out.close();

            // Commit session
            Intent commitIntent = new Intent(context, InstallResultReceiver.class);
            android.app.PendingIntent pendingIntent = android.app.PendingIntent.getBroadcast(
                context, sessionId, commitIntent,
                Build.VERSION.SDK_INT >= Build.VERSION_CODES.S ?
                    android.app.PendingIntent.FLAG_MUTABLE : 0);

            session.commit(pendingIntent.getIntentSender());
            session.close();

            String packageName = getPackageNameFromApk(apkFile);
            return new InstallResult(true, packageName, METHOD_SESSION_INSTALL,
                "Session submitted, check result via callback");

        } catch (Exception e) {
            Log.e(TAG, "PackageInstaller failed: " + e.getMessage());
            return new InstallResult(false, null, METHOD_SESSION_INSTALL, e.getMessage());
        }
    }

    /**
     * Install via user-guided interaction (fallback)
     */
    @NonNull
    private InstallResult installViaUserGuided(@NonNull File apkFile) {
        Log.i(TAG, "Installing via user-guided interaction");

        try {
            Intent installIntent = new Intent(Intent.ACTION_INSTALL_PACKAGE);
            installIntent.setData(Uri.fromFile(apkFile));
            installIntent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
            installIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.N) {
                // Use FileProvider for Android 7.0+
                Uri apkUri = getApkUri(apkFile);
                installIntent.setData(apkUri);
                installIntent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION);
            }

            context.startActivity(installIntent);

            return new InstallResult(false, null, METHOD_USER_GUIDED,
                "User interaction required - install dialog shown");
        } catch (Exception e) {
            Log.e(TAG, "User-guided install failed: " + e.getMessage());
            return new InstallResult(false, null, METHOD_USER_GUIDED, e.getMessage());
        }
    }

    /**
     * Uninstall package
     */
    @NonNull
    public InstallResult uninstall(@NonNull String packageName) {
        Log.i(TAG, "Uninstalling package: " + packageName);

        int preferredMethod = determineUninstallMethod();

        switch (preferredMethod) {
            case METHOD_PM_INSTALL:
                return uninstallViaPackageManager(packageName);

            case METHOD_ROOT_INSTALL:
                return uninstallViaRoot(packageName);

            default:
                return uninstallViaUserGuided(packageName);
        }
    }

    /**
     * Determine uninstall method
     */
    private int determineUninstallMethod() {
        if (permissionManager.checkPermission(SystemPermissionManager.PERMISSION_DELETE_PACKAGES)) {
            return METHOD_PM_INSTALL;
        }

        if (permissionManager.checkRootAccess()) {
            return METHOD_ROOT_INSTALL;
        }

        return METHOD_USER_GUIDED;
    }

    /**
     * Uninstall via package manager
     */
    @NonNull
    private InstallResult uninstallViaPackageManager(@NonNull String packageName) {
        try {
            String command = "pm uninstall " + packageName;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                Log.i(TAG, "Uninstall successful: " + packageName);
                return new InstallResult(true, packageName, METHOD_PM_INSTALL, "Success");
            } else {
                return new InstallResult(false, packageName, METHOD_PM_INSTALL, "PM uninstall failed");
            }
        } catch (Exception e) {
            Log.e(TAG, "PM uninstall failed: " + e.getMessage());
            return new InstallResult(false, packageName, METHOD_PM_INSTALL, e.getMessage());
        }
    }

    /**
     * Uninstall via root
     */
    @NonNull
    private InstallResult uninstallViaRoot(@NonNull String packageName) {
        try {
            String command = "pm uninstall " + packageName;
            Process process = Runtime.getRuntime().exec(new String[]{"su", "-c", command});
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                Log.i(TAG, "Root uninstall successful: " + packageName);
                return new InstallResult(true, packageName, METHOD_ROOT_INSTALL, "Success");
            } else {
                return new InstallResult(false, packageName, METHOD_ROOT_INSTALL, "Root uninstall failed");
            }
        } catch (Exception e) {
            Log.e(TAG, "Root uninstall failed: " + e.getMessage());
            return new InstallResult(false, packageName, METHOD_ROOT_INSTALL, e.getMessage());
        }
    }

    /**
     * Uninstall via user-guided interaction
     */
    @NonNull
    private InstallResult uninstallViaUserGuided(@NonNull String packageName) {
        try {
            Intent uninstallIntent = new Intent(Intent.ACTION_UNINSTALL_PACKAGE);
            uninstallIntent.setData(Uri.parse("package:" + packageName));
            uninstallIntent.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
            context.startActivity(uninstallIntent);

            return new InstallResult(false, packageName, METHOD_USER_GUIDED,
                "User interaction required - uninstall dialog shown");
        } catch (Exception e) {
            Log.e(TAG, "User-guided uninstall failed: " + e.getMessage());
            return new InstallResult(false, packageName, METHOD_USER_GUIDED, e.getMessage());
        }
    }

    /**
     * Copy APK to temp location
     */
    @NonNull
    private File copyToTempLocation(@NonNull File apkFile) throws IOException {
        File tempDir = new File(context.getCacheDir(), "install");
        tempDir.mkdirs();
        File tempFile = new File(tempDir, apkFile.getName());

        FileInputStream in = new FileInputStream(apkFile);
        FileOutputStream out = new FileOutputStream(tempFile);
        byte[] buffer = new byte[4096];
        int bytesRead;
        while ((bytesRead = in.read(buffer)) != -1) {
            out.write(buffer, 0, bytesRead);
        }
        in.close();
        out.close();

        return tempFile;
    }

    /**
     * Get package name from APK
     */
    @NonNull
    private String getPackageNameFromApk(@NonNull File apkFile) {
        try {
            android.content.pm.PackageInfo info =
                context.getPackageManager().getPackageArchiveInfo(apkFile.getAbsolutePath(), 0);
            if (info != null) {
                return info.packageName;
            }
        } catch (Exception e) {
            Log.w(TAG, "Failed to get package name from APK: " + e.getMessage());
        }
        return "unknown";
    }

    /**
     * Get APK URI for Android 7.0+
     */
    @NonNull
    private Uri getApkUri(@NonNull File apkFile) {
        // This would need a FileProvider implementation
        // For now, return file URI (will need adjustment for production)
        return Uri.fromFile(apkFile);
    }

    /**
     * Broadcast receiver for install results (inner class for documentation)
     */
    public static class InstallResultReceiver extends android.content.BroadcastReceiver {
        @Override
        public void onReceive(Context context, Intent intent) {
            int status = intent.getIntExtra(
                android.content.pm.PackageInstaller.EXTRA_STATUS,
                android.content.pm.PackageInstaller.STATUS_FAILURE);

            String packageName = intent.getData() != null ?
                intent.getData().getSchemeSpecificPart() : "unknown";

            Log.i("InstallResultReceiver", "Install result: " + status + " for " + packageName);
        }
    }
}