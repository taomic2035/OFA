package com.ofa.agent.automation;

import androidx.annotation.NonNull;

/**
 * Screen dimension information.
 */
public class ScreenDimension {

    private final int width;
    private final int height;
    private final int densityDpi;
    private final float scaledDensity;

    public ScreenDimension(int width, int height, int densityDpi, float scaledDensity) {
        this.width = width;
        this.height = height;
        this.densityDpi = densityDpi;
        this.scaledDensity = scaledDensity;
    }

    public int getWidth() {
        return width;
    }

    public int getHeight() {
        return height;
    }

    public int getDensityDpi() {
        return densityDpi;
    }

    public float getScaledDensity() {
        return scaledDensity;
    }

    /**
     * Convert dp to pixels
     */
    public int dpToPx(float dp) {
        return (int) (dp * scaledDensity + 0.5f);
    }

    /**
     * Convert pixels to dp
     */
    public float pxToDp(int px) {
        return px / scaledDensity;
    }

    /**
     * Get screen center X
     */
    public int getCenterX() {
        return width / 2;
    }

    /**
     * Get screen center Y
     */
    public int getCenterY() {
        return height / 2;
    }

    @NonNull
    @Override
    public String toString() {
        return "ScreenDimension{" + width + "x" + height + ", dpi=" + densityDpi + "}";
    }
}