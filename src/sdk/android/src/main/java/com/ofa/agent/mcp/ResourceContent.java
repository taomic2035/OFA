package com.ofa.agent.mcp;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

/**
 * Resource Content - content read from a resource.
 */
public class ResourceContent {

    private final String uri;
    private final String mimeType;
    private final byte[] data;
    private final String text;
    private final long lastModified;

    /**
     * Binary content constructor
     */
    public ResourceContent(@NonNull String uri, @NonNull String mimeType,
                           @NonNull byte[] data, long lastModified) {
        this.uri = uri;
        this.mimeType = mimeType;
        this.data = data;
        this.text = null;
        this.lastModified = lastModified;
    }

    /**
     * Text content constructor
     */
    public ResourceContent(@NonNull String uri, @NonNull String mimeType,
                           @NonNull String text, long lastModified) {
        this.uri = uri;
        this.mimeType = mimeType;
        this.data = null;
        this.text = text;
        this.lastModified = lastModified;
    }

    @NonNull
    public String getUri() {
        return uri;
    }

    @NonNull
    public String getMimeType() {
        return mimeType;
    }

    @Nullable
    public byte[] getBinaryData() {
        return data;
    }

    @Nullable
    public String getTextData() {
        return text;
    }

    public boolean isText() {
        return text != null;
    }

    public boolean isBinary() {
        return data != null;
    }

    public long getLastModified() {
        return lastModified;
    }

    /**
     * Get data as string (for text or convertible binary)
     */
    @Nullable
    public String asString() {
        if (text != null) return text;
        if (data != null) return new String(data);
        return null;
    }

    /**
     * Get data size
     */
    public int getSize() {
        if (data != null) return data.length;
        if (text != null) return text.length();
        return 0;
    }
}