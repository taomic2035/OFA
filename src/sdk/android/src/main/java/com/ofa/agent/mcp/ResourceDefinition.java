package com.ofa.agent.mcp;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import org.json.JSONObject;

/**
 * Resource Definition - describes a resource accessible via MCP.
 * Resources can be files, documents, or any readable content.
 */
public class ResourceDefinition {

    private final String uri;
    private final String name;
    private final String description;
    private final String mimeType;
    private final boolean readable;
    private final boolean writable;
    private final long size;

    /**
     * Create a resource definition
     * @param uri Resource URI (e.g., "file:///path/to/file", "content://contacts")
     * @param name Human-readable name
     * @param description Resource description
     * @param mimeType MIME type (e.g., "text/plain", "application/json")
     */
    public ResourceDefinition(@NonNull String uri, @NonNull String name,
                              @Nullable String description, @Nullable String mimeType) {
        this.uri = uri;
        this.name = name;
        this.description = description != null ? description : "";
        this.mimeType = mimeType != null ? mimeType : "application/octet-stream";
        this.readable = true;
        this.writable = false;
        this.size = -1;
    }

    /**
     * Full constructor
     */
    public ResourceDefinition(@NonNull String uri, @NonNull String name,
                              @Nullable String description, @Nullable String mimeType,
                              boolean readable, boolean writable, long size) {
        this.uri = uri;
        this.name = name;
        this.description = description != null ? description : "";
        this.mimeType = mimeType != null ? mimeType : "application/octet-stream";
        this.readable = readable;
        this.writable = writable;
        this.size = size;
    }

    @NonNull
    public String getUri() {
        return uri;
    }

    @NonNull
    public String getName() {
        return name;
    }

    @NonNull
    public String getDescription() {
        return description;
    }

    @NonNull
    public String getMimeType() {
        return mimeType;
    }

    public boolean isReadable() {
        return readable;
    }

    public boolean isWritable() {
        return writable;
    }

    public long getSize() {
        return size;
    }

    /**
     * Convert to JSON for MCP protocol
     */
    @NonNull
    public JSONObject toJson() {
        JSONObject json = new JSONObject();
        try {
            json.put("uri", uri);
            json.put("name", name);
            json.put("description", description);
            json.put("mimeType", mimeType);
            json.put("readable", readable);
            json.put("writable", writable);
            if (size >= 0) {
                json.put("size", size);
            }
        } catch (Exception e) {
            // JSON creation should not fail
        }
        return json;
    }

    @Nullable
    public static ResourceDefinition fromJson(@NonNull JSONObject json) {
        try {
            String uri = json.getString("uri");
            String name = json.getString("name");
            String description = json.optString("description", "");
            String mimeType = json.optString("mimeType", "application/octet-stream");
            boolean readable = json.optBoolean("readable", true);
            boolean writable = json.optBoolean("writable", false);
            long size = json.optLong("size", -1);

            return new ResourceDefinition(uri, name, description, mimeType, readable, writable, size);
        } catch (Exception e) {
            return null;
        }
    }

    @NonNull
    @Override
    public String toString() {
        return "ResourceDefinition{" + uri + "}";
    }
}