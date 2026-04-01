package com.ofa.agent.tool.builtin.data;

import android.Manifest;
import android.content.ContentResolver;
import android.content.Context;
import android.database.Cursor;
import android.net.Uri;
import android.os.Environment;
import android.provider.MediaStore;
import android.util.Log;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.core.content.ContextCompat;

import com.ofa.agent.mcp.MCPProtocol;
import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.tool.ExecutionContext;
import com.ofa.agent.tool.ToolExecutor;
import com.ofa.agent.tool.PermissionManager;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONArray;
import org.json.JSONObject;

import java.util.Map;

/**
 * Media Tool - query and manage media files.
 */
public class MediaTool implements ToolExecutor {

    private static final String TAG = "MediaTool";

    private final Context context;
    private final ContentResolver contentResolver;

    public MediaTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.contentResolver = context.getContentResolver();
    }

    @NonNull
    @Override
    public String getToolId() {
        return "media";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Query and manage media files";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "query");

        switch (operation.toLowerCase()) {
            case "query":
                return executeQuery(args, ctx);
            case "images":
                return executeQueryImages(args, ctx);
            case "videos":
                return executeQueryVideos(args, ctx);
            case "audio":
                return executeQueryAudio(args, ctx);
            case "count":
                return executeCount(args, ctx);
            default:
                return new ToolResult(getToolId(), "Unknown operation: " + operation);
        }
    }

    @Override
    public boolean isAvailable() {
        return contentResolver != null;
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
        return PermissionManager.getStoragePermissions();
    }

    @Override
    public boolean validateArgs(@NonNull Map<String, Object> args) {
        return true;
    }

    @Override
    public int getEstimatedTimeMs() {
        return 500;
    }

    // ===== Tool Definitions =====

    @NonNull
    public static ToolDefinition getQueryDefinition() {
        JSONObject props = new JSONObject();
        try {
            JSONObject operation = new JSONObject();
            operation.put("type", "string");
            operation.put("description", "Operation: 'query', 'images', 'videos', 'audio'");
            operation.put("default", "query");
            props.put("operation", operation);

            JSONObject limit = new JSONObject();
            limit.put("type", "integer");
            limit.put("description", "Maximum results");
            limit.put("default", 50);
            props.put("limit", limit);

            JSONObject offset = new JSONObject();
            offset.put("type", "integer");
            offset.put("description", "Offset for pagination");
            offset.put("default", 0);
            props.put("offset", offset);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("media.query", "Query media files",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getImagesDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'images'");
        return new ToolDefinition("media.images", "Query image files",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getVideosDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'videos'");
        return new ToolDefinition("media.videos", "Query video files",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getAudioDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'audio'");
        return new ToolDefinition("media.audio", "Query audio files",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeQuery(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        return executeQueryImages(args, ctx);
    }

    @NonNull
    private ToolResult executeQueryImages(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        int limit = getIntArg(args, "limit", 50);
        int offset = getIntArg(args, "offset", 0);

        try {
            JSONArray imagesArray = new JSONArray();

            Uri uri = MediaStore.Images.Media.EXTERNAL_CONTENT_URI;
            String[] projection = new String[]{
                    MediaStore.Images.Media._ID,
                    MediaStore.Images.Media.DISPLAY_NAME,
                    MediaStore.Images.Media.DATE_ADDED,
                    MediaStore.Images.Media.DATE_MODIFIED,
                    MediaStore.Images.Media.SIZE,
                    MediaStore.Images.Media.WIDTH,
                    MediaStore.Images.Media.HEIGHT,
                    MediaStore.Images.Media.MIME_TYPE
            };

            Cursor cursor = contentResolver.query(
                    uri,
                    projection,
                    null, null,
                    MediaStore.Images.Media.DATE_ADDED + " DESC LIMIT " + limit + " OFFSET " + offset);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject image = new JSONObject();

                    image.put("id", cursor.getLong(0));
                    image.put("name", cursor.getString(1));
                    image.put("dateAdded", cursor.getLong(2));
                    image.put("dateModified", cursor.getLong(3));
                    image.put("size", cursor.getLong(4));
                    image.put("width", cursor.getInt(5));
                    image.put("height", cursor.getInt(6));
                    image.put("mimeType", cursor.getString(7));

                    // Build URI
                    Uri contentUri = Uri.withAppendedPath(uri, String.valueOf(cursor.getLong(0)));
                    image.put("uri", contentUri.toString());

                    imagesArray.put(image);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("type", "images");
            output.put("count", imagesArray.length());
            output.put("media", imagesArray);

            return new ToolResult(getToolId(), output, 300);

        } catch (Exception e) {
            Log.e(TAG, "Query images failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeQueryVideos(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        int limit = getIntArg(args, "limit", 50);
        int offset = getIntArg(args, "offset", 0);

        try {
            JSONArray videosArray = new JSONArray();

            Uri uri = MediaStore.Video.Media.EXTERNAL_CONTENT_URI;
            String[] projection = new String[]{
                    MediaStore.Video.Media._ID,
                    MediaStore.Video.Media.DISPLAY_NAME,
                    MediaStore.Video.Media.DATE_ADDED,
                    MediaStore.Video.Media.SIZE,
                    MediaStore.Video.Media.DURATION,
                    MediaStore.Video.Media.WIDTH,
                    MediaStore.Video.Media.HEIGHT,
                    MediaStore.Video.Media.MIME_TYPE
            };

            Cursor cursor = contentResolver.query(
                    uri,
                    projection,
                    null, null,
                    MediaStore.Video.Media.DATE_ADDED + " DESC LIMIT " + limit + " OFFSET " + offset);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject video = new JSONObject();

                    video.put("id", cursor.getLong(0));
                    video.put("name", cursor.getString(1));
                    video.put("dateAdded", cursor.getLong(2));
                    video.put("size", cursor.getLong(3));
                    video.put("duration", cursor.getLong(4));
                    video.put("width", cursor.getInt(5));
                    video.put("height", cursor.getInt(6));
                    video.put("mimeType", cursor.getString(7));

                    Uri contentUri = Uri.withAppendedPath(uri, String.valueOf(cursor.getLong(0)));
                    video.put("uri", contentUri.toString());

                    videosArray.put(video);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("type", "videos");
            output.put("count", videosArray.length());
            output.put("media", videosArray);

            return new ToolResult(getToolId(), output, 300);

        } catch (Exception e) {
            Log.e(TAG, "Query videos failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeQueryAudio(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        int limit = getIntArg(args, "limit", 50);
        int offset = getIntArg(args, "offset", 0);

        try {
            JSONArray audioArray = new JSONArray();

            Uri uri = MediaStore.Audio.Media.EXTERNAL_CONTENT_URI;
            String[] projection = new String[]{
                    MediaStore.Audio.Media._ID,
                    MediaStore.Audio.Media.DISPLAY_NAME,
                    MediaStore.Audio.Media.TITLE,
                    MediaStore.Audio.Media.ARTIST,
                    MediaStore.Audio.Media.ALBUM,
                    MediaStore.Audio.Media.DURATION,
                    MediaStore.Audio.Media.SIZE,
                    MediaStore.Audio.Media.MIME_TYPE
            };

            Cursor cursor = contentResolver.query(
                    uri,
                    projection,
                    null, null,
                    MediaStore.Audio.Media.DATE_ADDED + " DESC LIMIT " + limit + " OFFSET " + offset);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject audio = new JSONObject();

                    audio.put("id", cursor.getLong(0));
                    audio.put("name", cursor.getString(1));
                    audio.put("title", cursor.getString(2));
                    audio.put("artist", cursor.getString(3));
                    audio.put("album", cursor.getString(4));
                    audio.put("duration", cursor.getLong(5));
                    audio.put("size", cursor.getLong(6));
                    audio.put("mimeType", cursor.getString(7));

                    Uri contentUri = Uri.withAppendedPath(uri, String.valueOf(cursor.getLong(0)));
                    audio.put("uri", contentUri.toString());

                    audioArray.put(audio);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("type", "audio");
            output.put("count", audioArray.length());
            output.put("media", audioArray);

            return new ToolResult(getToolId(), output, 300);

        } catch (Exception e) {
            Log.e(TAG, "Query audio failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeCount(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String type = getStringArg(args, "type", "all");

        try {
            JSONObject output = new JSONObject();
            output.put("success", true);

            if ("all".equalsIgnoreCase(type) || "images".equalsIgnoreCase(type)) {
                Cursor cursor = contentResolver.query(
                        MediaStore.Images.Media.EXTERNAL_CONTENT_URI,
                        null, null, null, null);
                output.put("images", cursor != null ? cursor.getCount() : 0);
                if (cursor != null) cursor.close();
            }

            if ("all".equalsIgnoreCase(type) || "videos".equalsIgnoreCase(type)) {
                Cursor cursor = contentResolver.query(
                        MediaStore.Video.Media.EXTERNAL_CONTENT_URI,
                        null, null, null, null);
                output.put("videos", cursor != null ? cursor.getCount() : 0);
                if (cursor != null) cursor.close();
            }

            if ("all".equalsIgnoreCase(type) || "audio".equalsIgnoreCase(type)) {
                Cursor cursor = contentResolver.query(
                        MediaStore.Audio.Media.EXTERNAL_CONTENT_URI,
                        null, null, null, null);
                output.put("audio", cursor != null ? cursor.getCount() : 0);
                if (cursor != null) cursor.close();
            }

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "Count media failed", e);
            return new ToolResult(getToolId(), "Count failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

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
        } catch (NumberFormatException e) {
            return defaultVal;
        }
    }
}