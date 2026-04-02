package com.ofa.agent.tool;

import android.content.Context;
import android.util.Log;

import androidx.annotation.NonNull;

import com.ofa.agent.mcp.ToolDefinition;
import com.ofa.agent.automation.UITool;
import com.ofa.agent.tool.builtin.ai.SpeechTool;
import com.ofa.agent.tool.builtin.data.CalendarTool;
import com.ofa.agent.tool.builtin.data.ContactsTool;
import com.ofa.agent.tool.builtin.data.MediaTool;
import com.ofa.agent.tool.builtin.device.BatteryTool;
import com.ofa.agent.tool.builtin.device.BluetoothTool;
import com.ofa.agent.tool.builtin.device.CameraTool;
import com.ofa.agent.tool.builtin.device.NFCTool;
import com.ofa.agent.tool.builtin.device.SensorTool;
import com.ofa.agent.tool.builtin.device.WiFiTool;
import com.ofa.agent.tool.builtin.system.AppTool;
import com.ofa.agent.tool.builtin.system.ClipboardTool;
import com.ofa.agent.tool.builtin.system.FileTool;
import com.ofa.agent.tool.builtin.system.NotificationTool;
import com.ofa.agent.tool.builtin.system.SettingsTool;

/**
 * Built-in Tools - factory class for registering all built-in tools.
 */
public class BuiltInTools {

    private static final String TAG = "BuiltInTools";

    /**
     * Register all built-in tools with a ToolRegistry
     * @param context Android context
     * @param registry Tool registry to register tools with
     */
    public static void registerAll(@NonNull Context context, @NonNull ToolRegistry registry) {
        Log.i(TAG, "Registering all built-in tools...");

        // System tools
        registerSystemTools(context, registry);

        // Device tools
        registerDeviceTools(context, registry);

        // Data tools
        registerDataTools(context, registry);

        // AI tools
        registerAITools(context, registry);

        // Automation tools (UI automation)
        registerAutomationTools(context, registry);

        Log.i(TAG, "Registered " + registry.getCount() + " built-in tools");
    }

    /**
     * Register system tools
     */
    public static void registerSystemTools(@NonNull Context context, @NonNull ToolRegistry registry) {
        // App Tool
        AppTool appTool = new AppTool(context);
        registry.register(AppTool.getLaunchDefinition(), appTool);
        registry.register(AppTool.getListDefinition(), appTool);
        registry.register(AppTool.getInfoDefinition(), appTool);

        // Settings Tool
        SettingsTool settingsTool = new SettingsTool(context);
        registry.register(SettingsTool.getGetDefinition(), settingsTool);
        registry.register(SettingsTool.getSetDefinition(), settingsTool);

        // Clipboard Tool
        ClipboardTool clipboardTool = new ClipboardTool(context);
        registry.register(ClipboardTool.getReadDefinition(), clipboardTool);
        registry.register(ClipboardTool.getWriteDefinition(), clipboardTool);
        registry.register(ClipboardTool.getClearDefinition(), clipboardTool);

        // File Tool
        FileTool fileTool = new FileTool(context);
        registry.register(FileTool.getReadDefinition(), fileTool);
        registry.register(FileTool.getWriteDefinition(), fileTool);
        registry.register(FileTool.getListDefinition(), fileTool);

        // Notification Tool
        NotificationTool notificationTool = new NotificationTool(context);
        registry.register(NotificationTool.getSendDefinition(), notificationTool);
        registry.register(NotificationTool.getCancelDefinition(), notificationTool);

        Log.d(TAG, "Registered system tools");
    }

    /**
     * Register device tools
     */
    public static void registerDeviceTools(@NonNull Context context, @NonNull ToolRegistry registry) {
        // Camera Tool
        CameraTool cameraTool = new CameraTool(context);
        registry.register(CameraTool.getCaptureDefinition(), cameraTool);
        registry.register(CameraTool.getScanDefinition(), cameraTool);
        registry.register(CameraTool.getListDefinition(), cameraTool);

        // Bluetooth Tool
        BluetoothTool bluetoothTool = new BluetoothTool(context);
        registry.register(BluetoothTool.getScanDefinition(), bluetoothTool);
        registry.register(BluetoothTool.getListDefinition(), bluetoothTool);
        registry.register(BluetoothTool.getInfoDefinition(), bluetoothTool);

        // WiFi Tool
        WiFiTool wifiTool = new WiFiTool(context);
        registry.register(WiFiTool.getScanDefinition(), wifiTool);
        registry.register(WiFiTool.getListDefinition(), wifiTool);
        registry.register(WiFiTool.getInfoDefinition(), wifiTool);

        // NFC Tool
        NFCTool nfcTool = new NFCTool(context);
        registry.register(NFCTool.getStatusDefinition(), nfcTool);
        registry.register(NFCTool.getReadDefinition(), nfcTool);
        registry.register(NFCTool.getWriteDefinition(), nfcTool);

        // Sensor Tool
        SensorTool sensorTool = new SensorTool(context);
        registry.register(SensorTool.getListDefinition(), sensorTool);
        registry.register(SensorTool.getReadDefinition(), sensorTool);
        registry.register(SensorTool.getInfoDefinition(), sensorTool);

        // Battery Tool
        BatteryTool batteryTool = new BatteryTool(context);
        registry.register(BatteryTool.getStatusDefinition(), batteryTool);

        Log.d(TAG, "Registered device tools");
    }

    /**
     * Register data tools
     */
    public static void registerDataTools(@NonNull Context context, @NonNull ToolRegistry registry) {
        // Contacts Tool
        ContactsTool contactsTool = new ContactsTool(context);
        registry.register(ContactsTool.getQueryDefinition(), contactsTool);
        registry.register(ContactsTool.getSearchDefinition(), contactsTool);
        registry.register(ContactsTool.getCountDefinition(), contactsTool);

        // Calendar Tool
        CalendarTool calendarTool = new CalendarTool(context);
        registry.register(CalendarTool.getQueryDefinition(), calendarTool);
        registry.register(CalendarTool.getCalendarsDefinition(), calendarTool);
        registry.register(CalendarTool.getTodayDefinition(), calendarTool);

        // Media Tool
        MediaTool mediaTool = new MediaTool(context);
        registry.register(MediaTool.getQueryDefinition(), mediaTool);
        registry.register(MediaTool.getImagesDefinition(), mediaTool);
        registry.register(MediaTool.getVideosDefinition(), mediaTool);
        registry.register(MediaTool.getAudioDefinition(), mediaTool);

        Log.d(TAG, "Registered data tools");
    }

    /**
     * Register AI tools
     */
    public static void registerAITools(@NonNull Context context, @NonNull ToolRegistry registry) {
        // Speech Tool
        SpeechTool speechTool = new SpeechTool(context);
        registry.register(SpeechTool.getSpeakDefinition(), speechTool);
        registry.register(SpeechTool.getStopDefinition(), speechTool);
        registry.register(SpeechTool.getStatusDefinition(), speechTool);

        Log.d(TAG, "Registered AI tools");
    }

    /**
     * Register automation tools (UI automation)
     */
    public static void registerAutomationTools(@NonNull Context context, @NonNull ToolRegistry registry) {
        // UI Tool - provides UI automation operations
        UITool uiTool = new UITool(context);

        // Phase 1 tools
        registry.register(UITool.getClickDefinition(), uiTool);
        registry.register(UITool.getLongClickDefinition(), uiTool);
        registry.register(UITool.getSwipeDefinition(), uiTool);
        registry.register(UITool.getInputDefinition(), uiTool);
        registry.register(UITool.getFindDefinition(), uiTool);
        registry.register(UITool.getWaitDefinition(), uiTool);
        registry.register(UITool.getScrollFindDefinition(), uiTool);

        // Phase 2 tools
        registry.register(UITool.getPullToRefreshDefinition(), uiTool);
        registry.register(UITool.getCaptureDefinition(), uiTool);
        registry.register(UITool.getWaitForStableDefinition(), uiTool);
        registry.register(UITool.getStartRecordDefinition(), uiTool);
        registry.register(UITool.getStopRecordDefinition(), uiTool);
        registry.register(UITool.getReplayDefinition(), uiTool);
        registry.register(UITool.getFindTextDefinition(), uiTool);

        Log.d(TAG, "Registered automation tools");
    }

    /**
     * Get list of all tool definitions (for documentation/discovery)
     */
    @NonNull
    public static ToolDefinition[] getAllDefinitions(@NonNull Context context) {
        return new ToolDefinition[] {
            // System
            AppTool.getLaunchDefinition(),
            AppTool.getListDefinition(),
            AppTool.getInfoDefinition(),
            SettingsTool.getGetDefinition(),
            SettingsTool.getSetDefinition(),
            ClipboardTool.getReadDefinition(),
            ClipboardTool.getWriteDefinition(),
            FileTool.getReadDefinition(),
            FileTool.getWriteDefinition(),
            FileTool.getListDefinition(),
            NotificationTool.getSendDefinition(),

            // Device
            CameraTool.getCaptureDefinition(),
            CameraTool.getScanDefinition(),
            CameraTool.getListDefinition(),
            BluetoothTool.getScanDefinition(),
            BluetoothTool.getListDefinition(),
            WiFiTool.getScanDefinition(),
            WiFiTool.getListDefinition(),
            NFCTool.getStatusDefinition(),
            SensorTool.getListDefinition(),
            SensorTool.getReadDefinition(),
            BatteryTool.getStatusDefinition(),

            // Data
            ContactsTool.getQueryDefinition(),
            ContactsTool.getSearchDefinition(),
            CalendarTool.getQueryDefinition(),
            CalendarTool.getTodayDefinition(),
            MediaTool.getImagesDefinition(),
            MediaTool.getVideosDefinition(),

            // AI
            SpeechTool.getSpeakDefinition(),

            // Automation (Phase 1)
            UITool.getClickDefinition(),
            UITool.getSwipeDefinition(),
            UITool.getInputDefinition(),
            UITool.getFindDefinition(),
            UITool.getWaitDefinition(),
            UITool.getScrollFindDefinition(),

            // Automation (Phase 2)
            UITool.getPullToRefreshDefinition(),
            UITool.getCaptureDefinition(),
            UITool.getWaitForStableDefinition(),
            UITool.getStartRecordDefinition(),
            UITool.getStopRecordDefinition(),
            UITool.getReplayDefinition(),
            UITool.getFindTextDefinition()
        };
    }
}