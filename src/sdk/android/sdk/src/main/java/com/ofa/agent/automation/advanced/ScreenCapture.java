package com.ofa.agent.automation.advanced;

import android.content.Context;
import android.graphics.Bitmap;
import android.graphics.PixelFormat;
import android.hardware.display.DisplayManager;
import android.hardware.display.VirtualDisplay;
import android.media.Image;
import android.media.ImageReader;
import android.media.projection.MediaProjection;
import android.media.projection.MediaProjectionManager;
import android.os.Build;
import android.os.Handler;
import android.os.Looper;
import android.util.DisplayMetrics;
import android.util.Log;
import android.view.WindowManager;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;
import androidx.annotation.RequiresApi;

import com.ofa.agent.automation.ScreenDimension;

import java.io.File;
import java.io.FileOutputStream;
import java.io.IOException;
import java.nio.ByteBuffer;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.Locale;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicReference;

/**
 * Screen Capture - captures screenshots using MediaProjection API.
 * Requires user permission through MediaProjection.
 */
public class ScreenCapture {

    private static final String TAG = "ScreenCapture";

    private final Context context;
    private final WindowManager windowManager;
    private final MediaProjectionManager projectionManager;
    private final ScreenDimension screenDimension;

    private MediaProjection mediaProjection;
    private VirtualDisplay virtualDisplay;
    private ImageReader imageReader;

    private volatile boolean initialized = false;
    private int resultCode;
    private android.content.Intent resultData;

    public ScreenCapture(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.windowManager = (WindowManager) context.getSystemService(Context.WINDOW_SERVICE);
        this.projectionManager = (MediaProjectionManager)
            context.getSystemService(Context.MEDIA_PROJECTION_SERVICE);

        DisplayMetrics metrics = new DisplayMetrics();
        windowManager.getDefaultDisplay().getRealMetrics(metrics);
        this.screenDimension = new ScreenDimension(
            metrics.widthPixels,
            metrics.heightPixels,
            metrics.densityDpi,
            metrics.scaledDensity
        );
    }

    /**
     * Initialize with MediaProjection result
     * Call this from onActivityResult after requesting screen capture permission
     */
    public void initialize(int resultCode, @NonNull android.content.Intent resultData) {
        this.resultCode = resultCode;
        this.resultData = resultData;
        this.initialized = true;
        Log.i(TAG, "ScreenCapture initialized");
    }

    /**
     * Check if screen capture is available
     */
    public boolean isAvailable() {
        return initialized && resultData != null;
    }

    /**
     * Request screen capture permission
     * This will show a system dialog asking user for permission
     */
    @RequiresApi(api = Build.VERSION_CODES.LOLLIPOP)
    public void requestPermission(@NonNull android.app.Activity activity, int requestCode) {
        activity.startActivityForResult(
            projectionManager.createScreenCaptureIntent(),
            requestCode
        );
    }

    /**
     * Take screenshot and return bitmap
     */
    @Nullable
    public Bitmap captureBitmap() {
        if (!isAvailable()) {
            Log.w(TAG, "ScreenCapture not initialized");
            return null;
        }

        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.LOLLIPOP) {
            Log.w(TAG, "Screen capture requires API 21+");
            return null;
        }

        MediaProjection projection = null;
        VirtualDisplay display = null;
        ImageReader reader = null;

        try {
            projection = projectionManager.getMediaProjection(resultCode, resultData);
            if (projection == null) {
                Log.e(TAG, "Failed to create MediaProjection");
                return null;
            }

            int width = screenDimension.getWidth();
            int height = screenDimension.getHeight();
            int density = screenDimension.getDensityDpi();

            reader = ImageReader.newInstance(width, height, PixelFormat.RGBA_8888, 2);

            display = projection.createVirtualDisplay(
                "ScreenCapture",
                width, height, density,
                DisplayManager.VIRTUAL_DISPLAY_FLAG_AUTO_MIRROR,
                reader.getSurface(),
                null,
                new Handler(Looper.getMainLooper())
            );

            // Wait for image
            final CountDownLatch latch = new CountDownLatch(1);
            final AtomicReference<Image> imageRef = new AtomicReference<>();

            reader.setOnImageAvailableListener(readerInstance -> {
                Image image = readerInstance.acquireLatestImage();
                if (image != null && imageRef.compareAndSet(null, image)) {
                    latch.countDown();
                }
            }, new Handler(Looper.getMainLooper()));

            // Wait with timeout
            if (!latch.await(3, TimeUnit.SECONDS)) {
                Log.w(TAG, "Timeout waiting for screenshot");
                return null;
            }

            Image image = imageRef.get();
            if (image == null) {
                Log.w(TAG, "No image captured");
                return null;
            }

            // Convert to bitmap
            Bitmap bitmap = imageToBitmap(image);
            image.close();

            return bitmap;

        } catch (Exception e) {
            Log.e(TAG, "Error capturing screenshot", e);
            return null;
        } finally {
            if (display != null) {
                display.release();
            }
            if (reader != null) {
                reader.setOnImageAvailableListener(null, null);
            }
            if (projection != null) {
                projection.stop();
            }
        }
    }

    /**
     * Take screenshot and save to file
     */
    @Nullable
    public String captureToFile(@Nullable File directory) {
        Bitmap bitmap = captureBitmap();
        if (bitmap == null) {
            return null;
        }

        try {
            File outputDir = directory != null ? directory : context.getCacheDir();
            String fileName = "screenshot_" +
                new SimpleDateFormat("yyyyMMdd_HHmmss", Locale.getDefault()).format(new Date()) +
                ".png";
            File outputFile = new File(outputDir, fileName);

            try (FileOutputStream fos = new FileOutputStream(outputFile)) {
                bitmap.compress(Bitmap.CompressFormat.PNG, 100, fos);
            }

            Log.i(TAG, "Screenshot saved to: " + outputFile.getAbsolutePath());
            return outputFile.getAbsolutePath();

        } catch (IOException e) {
            Log.e(TAG, "Error saving screenshot", e);
            return null;
        } finally {
            bitmap.recycle();
        }
    }

    /**
     * Take screenshot to default location
     */
    @Nullable
    public String captureToDefaultLocation() {
        return captureToFile(null);
    }

    /**
     * Capture region of screen
     */
    @Nullable
    public Bitmap captureRegion(int x, int y, int width, int height) {
        Bitmap fullBitmap = captureBitmap();
        if (fullBitmap == null) {
            return null;
        }

        try {
            // Validate bounds
            if (x < 0) x = 0;
            if (y < 0) y = 0;
            if (x + width > fullBitmap.getWidth()) {
                width = fullBitmap.getWidth() - x;
            }
            if (y + height > fullBitmap.getHeight()) {
                height = fullBitmap.getHeight() - y;
            }

            return Bitmap.createBitmap(fullBitmap, x, y, width, height);
        } finally {
            fullBitmap.recycle();
        }
    }

    /**
     * Compare two bitmaps for similarity
     * @return Similarity percentage (0.0 - 1.0)
     */
    public static float compareBitmaps(@NonNull Bitmap bitmap1, @NonNull Bitmap bitmap2) {
        if (bitmap1.getWidth() != bitmap2.getWidth() ||
            bitmap1.getHeight() != bitmap2.getHeight()) {
            // Resize for comparison
            int width = Math.min(bitmap1.getWidth(), bitmap2.getWidth());
            int height = Math.min(bitmap1.getHeight(), bitmap2.getHeight());

            Bitmap resized1 = Bitmap.createScaledBitmap(bitmap1, width, height, false);
            Bitmap resized2 = Bitmap.createScaledBitmap(bitmap2, width, height, false);

            float similarity = calculateSimilarity(resized1, resized2);

            resized1.recycle();
            resized2.recycle();

            return similarity;
        }

        return calculateSimilarity(bitmap1, bitmap2);
    }

    private static float calculateSimilarity(@NonNull Bitmap bitmap1, @NonNull Bitmap bitmap2) {
        int width = bitmap1.getWidth();
        int height = bitmap1.getHeight();
        long diff = 0;
        long total = (long) width * height * 3 * 255; // Max possible diff

        for (int y = 0; y < height; y++) {
            for (int x = 0; x < width; x++) {
                int pixel1 = bitmap1.getPixel(x, y);
                int pixel2 = bitmap2.getPixel(x, y);

                int r1 = (pixel1 >> 16) & 0xFF;
                int g1 = (pixel1 >> 8) & 0xFF;
                int b1 = pixel1 & 0xFF;

                int r2 = (pixel2 >> 16) & 0xFF;
                int g2 = (pixel2 >> 8) & 0xFF;
                int b2 = pixel2 & 0xFF;

                diff += Math.abs(r1 - r2) + Math.abs(g1 - g2) + Math.abs(b1 - b2);
            }
        }

        return 1.0f - ((float) diff / total);
    }

    /**
     * Convert Image to Bitmap
     */
    @Nullable
    private Bitmap imageToBitmap(@NonNull Image image) {
        Image.Plane[] planes = image.getPlanes();
        if (planes == null || planes.length == 0) {
            return null;
        }

        Image.Plane plane = planes[0];
        ByteBuffer buffer = plane.getBuffer();
        int pixelStride = plane.getPixelStride();
        int rowStride = plane.getRowStride();
        int rowPadding = rowStride - pixelStride * screenDimension.getWidth();

        // Create bitmap
        Bitmap bitmap = Bitmap.createBitmap(
            screenDimension.getWidth() + rowPadding / pixelStride,
            screenDimension.getHeight(),
            Bitmap.Config.ARGB_8888
        );
        bitmap.copyPixelsFromBuffer(buffer);

        // Crop to actual size
        Bitmap croppedBitmap = Bitmap.createBitmap(
            bitmap, 0, 0,
            screenDimension.getWidth(),
            screenDimension.getHeight()
        );
        bitmap.recycle();

        return croppedBitmap;
    }

    /**
     * Release resources
     */
    public void release() {
        if (virtualDisplay != null) {
            virtualDisplay.release();
            virtualDisplay = null;
        }
        if (imageReader != null) {
            imageReader.setOnImageAvailableListener(null, null);
            imageReader = null;
        }
        if (mediaProjection != null) {
            mediaProjection.stop();
            mediaProjection = null;
        }
        Log.i(TAG, "ScreenCapture released");
    }
}