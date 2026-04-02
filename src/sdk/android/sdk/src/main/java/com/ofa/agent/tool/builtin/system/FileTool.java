package com.ofa.agent.tool.builtin.system;

import android.content.Context;
import android.os.Environment;
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

import java.io.File;
import java.io.FileInputStream;
import java.io.FileOutputStream;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * File Tool - read, write, list, delete files.
 * Works with app-private storage by default.
 */
public class FileTool implements ToolExecutor {

    private static final String TAG = "FileTool";
    private static final int MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB limit

    private final Context context;
    private final File filesDir;
    private final File cacheDir;

    public FileTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.filesDir = context.getFilesDir();
        this.cacheDir = context.getCacheDir();
    }

    @NonNull
    @Override
    public String getToolId() {
        return "file";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Read, write, list, delete files in app storage";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "list");

        switch (operation.toLowerCase()) {
            case "read":
                return executeRead(args, ctx);
            case "write":
                return executeWrite(args, ctx);
            case "list":
                return executeList(args, ctx);
            case "delete":
                return executeDelete(args, ctx);
            case "info":
                return executeInfo(args, ctx);
            case "exists":
                return executeExists(args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return filesDir != null && filesDir.exists();
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
        return null; // App-private storage needs no permissions
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        String operation = getStringArg(args, "operation", null);
        if (operation == null) return false;

        switch (operation.toLowerCase()) {
            case "read":
            case "delete":
            case "info":
            case "exists":
                return args.containsKey("path");
            case "write":
                return args.containsKey("path") && args.containsKey("content");
            case "list":
                return true;
            default:
                return false;
        }
    }

    @Override
    public int getEstimatedTimeMs() {
        return 200;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getReadDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'read'");
            props.put("operation", operation);

            JSONObject path = new JSONObject();
            path.put("type", "string");
            path.put("description", "File path relative to app files directory");
            props.put("path", path);

            JSONObject encoding = new JSONObject();
            encoding.put("type", "string");
            encoding.put("description", "Text encoding: utf-8, base64");
            encoding.put("default", "utf-8");
            props.put("encoding", encoding);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"path"});
        return new ToolDefinition("file.read", "Read file content",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getWriteDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'write'");
            props.put("operation", operation);

            JSONObject path = new JSONObject();
            path.put("type", "string");
            path.put("description", "File path relative to app files directory");
            props.put("path", path);

            JSONObject content = new JSONObject();
            content.put("type", "string");
            content.put("description", "Content to write");
            props.put("content", content);

            JSONObject encoding = new JSONObject();
            encoding.put("type", "string");
            encoding.put("description", "Content encoding: utf-8, base64");
            encoding.put("default", "utf-8");
            props.put("encoding", encoding);

            JSONObject append = new JSONObject();
            append.put("type", "boolean");
            append.put("description", "Append to existing file");
            append.put("default", false);
            props.put("append", append);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, new String[]{"path", "content"});
        return new ToolDefinition("file.write", "Write content to file",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getListDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'list'");
            props.put("operation", operation);

            JSONObject path = new JSONObject();
            path.put("type", "string");
            path.put("description", "Directory path (empty for root)");
            path.put("default", "");
            props.put("path", path);

            JSONObject recursive = new JSONObject();
            recursive.put("type", "boolean");
            recursive.put("description", "List recursively");
            recursive.put("default", false);
            props.put("recursive", recursive);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("file.list", "List files in directory",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeRead(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String path = getStringArg(args, "path", null);
        if (path == null) {
            return new ToolResult(getToolId(), "Missing path parameter");
        }

        String encoding = getStringArg(args, "encoding", "utf-8");

        try {
            File file = resolvePath(path);
            if (!file.exists()) {
                return new ToolResult(getToolId(), "File not found: " + path);
            }

            if (file.length() > MAX_FILE_SIZE) {
                return new ToolResult(getToolId(), "File too large (max " + MAX_FILE_SIZE + " bytes)");
            }

            FileInputStream fis = new FileInputStream(file);
            byte[] data = new byte[(int) file.length()];
            fis.read(data);
            fis.close();

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("path", path);
            output.put("size", file.length());

            if ("base64".equalsIgnoreCase(encoding)) {
                output.put("content", android.util.Base64.encodeToString(data, android.util.Base64.DEFAULT));
                output.put("encoding", "base64");
            } else {
                output.put("content", new String(data, StandardCharsets.UTF_8));
                output.put("encoding", "utf-8");
            }

            return new ToolResult(getToolId(), output, 50);

        } catch (Exception e) {
            Log.e(TAG, "Read file failed", e);
            return new ToolResult(getToolId(), "Read failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeWrite(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String path = getStringArg(args, "path", null);
        String content = getStringArg(args, "content", "");
        String encoding = getStringArg(args, "encoding", "utf-8");
        boolean append = getBooleanArg(args, "append", false);

        if (path == null) {
            return new ToolResult(getToolId(), "Missing path parameter");
        }

        try {
            File file = resolvePath(path);

            // Create parent directories if needed
            File parent = file.getParentFile();
            if (parent != null && !parent.exists()) {
                parent.mkdirs();
            }

            byte[] data;
            if ("base64".equalsIgnoreCase(encoding)) {
                data = android.util.Base64.decode(content, android.util.Base64.DEFAULT);
            } else {
                data = content.getBytes(StandardCharsets.UTF_8);
            }

            FileOutputStream fos = new FileOutputStream(file, append);
            fos.write(data);
            fos.close();

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("path", path);
            output.put("size", data.length);
            output.put("append", append);
            output.put("written", true);

            return new ToolResult(getToolId(), output, 50);

        } catch (Exception e) {
            Log.e(TAG, "Write file failed", e);
            return new ToolResult(getToolId(), "Write failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeList(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String path = getStringArg(args, "path", "");
        boolean recursive = getBooleanArg(args, "recursive", false);

        try {
            File dir = resolvePath(path);
            if (!dir.exists() || !dir.isDirectory()) {
                return new ToolResult(getToolId(), "Directory not found: " + path);
            }

            List<FileInfo> files = listFiles(dir, recursive);

            JSONArray filesArray = new JSONArray();
            for (FileInfo fi : files) {
                JSONObject fileJson = new JSONObject();
                fileJson.put("name", fi.name);
                fileJson.put("path", fi.relativePath);
                fileJson.put("size", fi.size);
                fileJson.put("isDirectory", fi.isDirectory);
                fileJson.put("lastModified", fi.lastModified);
                filesArray.put(fileJson);
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("path", path);
            output.put("count", files.size());
            output.put("files", filesArray);
            output.put("recursive", recursive);

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "List files failed", e);
            return new ToolResult(getToolId(), "List failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeDelete(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String path = getStringArg(args, "path", null);
        if (path == null) {
            return new ToolResult(getToolId(), "Missing path parameter");
        }

        try {
            File file = resolvePath(path);
            if (!file.exists()) {
                return new ToolResult(getToolId(), "File not found: " + path);
            }

            boolean deleted;
            if (file.isDirectory()) {
                deleted = deleteDirectory(file);
            } else {
                deleted = file.delete();
            }

            JSONObject output = new JSONObject();
            output.put("success", deleted);
            output.put("path", path);
            output.put("deleted", deleted);

            return new ToolResult(getToolId(), output, 50);

        } catch (Exception e) {
            Log.e(TAG, "Delete failed", e);
            return new ToolResult(getToolId(), "Delete failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeInfo(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String path = getStringArg(args, "path", null);
        if (path == null) {
            return new ToolResult(getToolId(), "Missing path parameter");
        }

        try {
            File file = resolvePath(path);

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("path", path);
            output.put("exists", file.exists());

            if (file.exists()) {
                output.put("isDirectory", file.isDirectory());
                output.put("isFile", file.isFile());
                output.put("size", file.length());
                output.put("lastModified", file.lastModified());
                output.put("canRead", file.canRead());
                output.put("canWrite", file.canWrite());
            }

            return new ToolResult(getToolId(), output, 20);

        } catch (Exception e) {
            Log.e(TAG, "Get info failed", e);
            return new ToolResult(getToolId(), "Get info failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeExists(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String path = getStringArg(args, "path", null);
        if (path == null) {
            return new ToolResult(getToolId(), "Missing path parameter");
        }

        File file = resolvePath(path);

        try {
            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("path", path);
            output.put("exists", file.exists());
            return new ToolResult(getToolId(), output, 10);
        } catch (Exception e) {
            return new ToolResult(getToolId(), "Error: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    @NonNull
    private File resolvePath(@NonNull String path) {
        if (path.startsWith("/")) {
            path = path.substring(1);
        }

        // Handle special paths
        if (path.equals("") || path.equals(".")) {
            return filesDir;
        }

        if (path.startsWith("cache/")) {
            return new File(cacheDir, path.substring(6));
        }

        return new File(filesDir, path);
    }

    @NonNull
    private List<FileInfo> listFiles(@NonNull File dir, boolean recursive) {
        List<FileInfo> result = new ArrayList<>();
        File[] files = dir.listFiles();

        if (files == null) return result;

        for (File file : files) {
            FileInfo fi = new FileInfo();
            fi.name = file.getName();
            fi.relativePath = getRelativePath(file);
            fi.size = file.length();
            fi.isDirectory = file.isDirectory();
            fi.lastModified = file.lastModified();

            result.add(fi);

            if (recursive && file.isDirectory()) {
                result.addAll(listFiles(file, true));
            }
        }

        return result;
    }

    @NonNull
    private String getRelativePath(@NonNull File file) {
        String absolute = file.getAbsolutePath();
        String base = filesDir.getAbsolutePath();
        if (absolute.startsWith(base)) {
            return absolute.substring(base.length() + 1);
        }
        return file.getName();
    }

    private boolean deleteDirectory(@NonNull File dir) {
        File[] files = dir.listFiles();
        if (files != null) {
            for (File file : files) {
                if (file.isDirectory()) {
                    deleteDirectory(file);
                } else {
                    file.delete();
                }
            }
        }
        return dir.delete();
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

    private static class FileInfo {
        String name;
        String relativePath;
        long size;
        boolean isDirectory;
        long lastModified;
    }
}