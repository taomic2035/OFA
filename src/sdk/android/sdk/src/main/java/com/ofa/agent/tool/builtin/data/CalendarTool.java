package com.ofa.agent.tool.builtin.data;

import android.Manifest;
import android.content.ContentResolver;
import android.content.Context;
import android.database.Cursor;
import android.provider.CalendarContract;
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
 * Calendar Tool - query and manage calendar events.
 */
public class CalendarTool implements ToolExecutor {

    private static final String TAG = "CalendarTool";

    private final Context context;
    private final ContentResolver contentResolver;

    public CalendarTool(@NonNull Context context) {
        this.context = context.getApplicationContext();
        this.contentResolver = context.getContentResolver();
    }

    @NonNull
    @Override
    public String getToolId() {
        return "calendar";
    }

    @NonNull
    @Override
    public String getDescription() {
        return "Query and manage calendar events";
    }

    @NonNull
    @Override
    public ToolResult execute(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        String operation = getStringArg(args, "operation", "query");

        switch (operation.toLowerCase()) {
            case "query":
                return executeQuery(args, ctx);
            case "calendars":
                return executeListCalendars(ctx);
            case "today":
                return executeToday(ctx);
            case "upcoming":
                return executeUpcoming(args, ctx);
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
        return PermissionManager.getCalendarPermissions();
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
            operation.put("description", "Operation: 'query'");
            operation.put("default", "query");
            props.put("operation", operation);

            JSONObject startDate = new JSONObject();
            startDate.put("type", "string");
            startDate.put("description", "Start date (ISO format)");
            props.put("startDate", startDate);

            JSONObject endDate = new JSONObject();
            endDate.put("type", "string");
            endDate.put("description", "End date (ISO format)");
            props.put("endDate", endDate);

            JSONObject limit = new JSONObject();
            limit.put("type", "integer");
            limit.put("description", "Maximum events to return");
            limit.put("default", 50);
            props.put("limit", limit);
        } catch (Exception e) {
            // Should not fail
        }

        JSONObject schema = MCPProtocol.buildObjectSchema(props, null);
        return new ToolDefinition("calendar.query", "Query calendar events",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getCalendarsDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'calendars'");
        return new ToolDefinition("calendar.calendars", "List available calendars",
                schema, true, null);
    }

    @NonNull
    public static ToolDefinition getTodayDefinition() {
        JSONObject schema = MCPProtocol.buildSimpleTextSchema("operation",
                "Operation: 'today'");
        return new ToolDefinition("calendar.today", "Get today's events",
                schema, true, null);
    }

    // ===== Internal Operations =====

    @NonNull
    private ToolResult executeQuery(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (!hasCalendarPermission()) {
            return new ToolResult(getToolId(), "Calendar permission not granted");
        }

        long startTime = getLongArg(args, "startDate", System.currentTimeMillis() - 24 * 60 * 60 * 1000);
        long endTime = getLongArg(args, "endDate", System.currentTimeMillis() + 7 * 24 * 60 * 60 * 1000);
        int limit = getIntArg(args, "limit", 50);

        try {
            JSONArray eventsArray = new JSONArray();

            String[] projection = new String[]{
                    CalendarContract.Events._ID,
                    CalendarContract.Events.TITLE,
                    CalendarContract.Events.DESCRIPTION,
                    CalendarContract.Events.DTSTART,
                    CalendarContract.Events.DTEND,
                    CalendarContract.Events.EVENT_LOCATION,
                    CalendarContract.Events.CALENDAR_ID
            };

            String selection = CalendarContract.Events.DTSTART + " >= ? AND " +
                    CalendarContract.Events.DTEND + " <= ?";
            String[] selectionArgs = new String[]{String.valueOf(startTime), String.valueOf(endTime)};

            Cursor cursor = contentResolver.query(
                    CalendarContract.Events.CONTENT_URI,
                    projection,
                    selection,
                    selectionArgs,
                    CalendarContract.Events.DTSTART + " ASC LIMIT " + limit);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject event = new JSONObject();

                    event.put("id", cursor.getLong(0));
                    event.put("title", cursor.getString(1));
                    event.put("description", cursor.getString(2));
                    event.put("startTime", cursor.getLong(3));
                    event.put("endTime", cursor.getLong(4));
                    event.put("location", cursor.getString(5));
                    event.put("calendarId", cursor.getLong(6));

                    eventsArray.put(event);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", eventsArray.length());
            output.put("events", eventsArray);
            output.put("queryStart", startTime);
            output.put("queryEnd", endTime);

            return new ToolResult(getToolId(), output, 300);

        } catch (Exception e) {
            Log.e(TAG, "Query calendar failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeListCalendars(@NonNull ExecutionContext ctx) {
        if (!hasCalendarPermission()) {
            return new ToolResult(getToolId(), "Calendar permission not granted");
        }

        try {
            JSONArray calendarsArray = new JSONArray();

            String[] projection = new String[]{
                    CalendarContract.Calendars._ID,
                    CalendarContract.Calendars.NAME,
                    CalendarContract.Calendars.ACCOUNT_NAME,
                    CalendarContract.Calendars.CALENDAR_DISPLAY_NAME,
                    CalendarContract.Calendars.CALENDAR_COLOR
            };

            Cursor cursor = contentResolver.query(
                    CalendarContract.Calendars.CONTENT_URI,
                    projection,
                    null, null, null);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject calendar = new JSONObject();

                    calendar.put("id", cursor.getLong(0));
                    calendar.put("name", cursor.getString(1));
                    calendar.put("accountName", cursor.getString(2));
                    calendar.put("displayName", cursor.getString(3));
                    calendar.put("color", cursor.getInt(4));

                    calendarsArray.put(calendar);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", calendarsArray.length());
            output.put("calendars", calendarsArray);

            return new ToolResult(getToolId(), output, 100);

        } catch (Exception e) {
            Log.e(TAG, "List calendars failed", e);
            return new ToolResult(getToolId(), "List failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeToday(@NonNull ExecutionContext ctx) {
        if (!hasCalendarPermission()) {
            return new ToolResult(getToolId(), "Calendar permission not granted");
        }

        long now = System.currentTimeMillis();
        long startOfDay = now - (now % (24 * 60 * 60 * 1000));
        long endOfDay = startOfDay + 24 * 60 * 60 * 1000;

        try {
            JSONArray eventsArray = new JSONArray();

            String[] projection = new String[]{
                    CalendarContract.Events._ID,
                    CalendarContract.Events.TITLE,
                    CalendarContract.Events.DESCRIPTION,
                    CalendarContract.Events.DTSTART,
                    CalendarContract.Events.DTEND,
                    CalendarContract.Events.EVENT_LOCATION
            };

            String selection = CalendarContract.Events.DTSTART + " >= ? AND " +
                    CalendarContract.Events.DTEND + " <= ?";
            String[] selectionArgs = new String[]{String.valueOf(startOfDay), String.valueOf(endOfDay)};

            Cursor cursor = contentResolver.query(
                    CalendarContract.Events.CONTENT_URI,
                    projection,
                    selection,
                    selectionArgs,
                    CalendarContract.Events.DTSTART + " ASC");

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject event = new JSONObject();

                    event.put("id", cursor.getLong(0));
                    event.put("title", cursor.getString(1));
                    event.put("description", cursor.getString(2));
                    event.put("startTime", cursor.getLong(3));
                    event.put("endTime", cursor.getLong(4));
                    event.put("location", cursor.getString(5));

                    eventsArray.put(event);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", eventsArray.length());
            output.put("events", eventsArray);
            output.put("date", startOfDay);

            return new ToolResult(getToolId(), output, 200);

        } catch (Exception e) {
            Log.e(TAG, "Get today's events failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    @NonNull
    private ToolResult executeUpcoming(@NonNull Map<String, Object> args, @NonNull ExecutionContext ctx) {
        if (!hasCalendarPermission()) {
            return new ToolResult(getToolId(), "Calendar permission not granted");
        }

        long now = System.currentTimeMillis();
        long endTime = now + 7 * 24 * 60 * 60 * 1000; // Next 7 days
        int limit = getIntArg(args, "limit", 20);

        try {
            JSONArray eventsArray = new JSONArray();

            String[] projection = new String[]{
                    CalendarContract.Events._ID,
                    CalendarContract.Events.TITLE,
                    CalendarContract.Events.DTSTART,
                    CalendarContract.Events.DTEND,
                    CalendarContract.Events.EVENT_LOCATION
            };

            String selection = CalendarContract.Events.DTSTART + " >= ? AND " +
                    CalendarContract.Events.DTEND + " <= ?";
            String[] selectionArgs = new String[]{String.valueOf(now), String.valueOf(endTime)};

            Cursor cursor = contentResolver.query(
                    CalendarContract.Events.CONTENT_URI,
                    projection,
                    selection,
                    selectionArgs,
                    CalendarContract.Events.DTSTART + " ASC LIMIT " + limit);

            if (cursor != null) {
                while (cursor.moveToNext()) {
                    JSONObject event = new JSONObject();

                    event.put("id", cursor.getLong(0));
                    event.put("title", cursor.getString(1));
                    event.put("startTime", cursor.getLong(2));
                    event.put("endTime", cursor.getLong(3));
                    event.put("location", cursor.getString(4));

                    // Calculate relative time
                    long start = cursor.getLong(2);
                    long diffHours = (start - now) / (60 * 60 * 1000);
                    event.put("hoursUntil", diffHours);

                    eventsArray.put(event);
                }
                cursor.close();
            }

            JSONObject output = new JSONObject();
            output.put("success", true);
            output.put("count", eventsArray.length());
            output.put("events", eventsArray);

            return new ToolResult(getToolId(), output, 200);

        } catch (Exception e) {
            Log.e(TAG, "Get upcoming events failed", e);
            return new ToolResult(getToolId(), "Query failed: " + e.getMessage());
        }
    }

    // ===== Helper Methods =====

    private boolean hasCalendarPermission() {
        return ContextCompat.checkSelfPermission(context, Manifest.permission.READ_CALENDAR)
                == android.content.pm.PackageManager.PERMISSION_GRANTED;
    }

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

    private long getLongArg(@NonNull Map<String, Object> args, @NonNull String key, long defaultVal) {
        Object value = args.get(key);
        if (value == null) return defaultVal;
        if (value instanceof Number) return ((Number) value).longValue();
        try {
            // Try parsing ISO date format
            return Long.parseLong(value.toString());
        } catch (NumberFormatException e) {
            return defaultVal;
        }
    }
}