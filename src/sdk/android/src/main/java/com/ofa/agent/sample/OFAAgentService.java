package com.ofa.agent.sample;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.Service;
import android.content.Context;
import android.content.Intent;
import android.os.IBinder;
import android.util.Log;

import androidx.annotation.Nullable;
import androidx.core.app.NotificationCompat;

import com.ofa.agent.OFAAgent;
import com.ofa.agent.R;
import com.ofa.agent.offline.OfflineLevel;
import com.ofa.agent.offline.OfflineManager;
import com.ofa.agent.skill.builtin.OfflineSkills;

/**
 * OFA Agent 后台服务示例
 *
 * 在后台运行 OFA Agent，保持与 Center 的连接
 */
public class OFAAgentService extends Service {

    private static final String TAG = "OFAAgentService";
    private static final String CHANNEL_ID = "ofa_agent_channel";
    private static final int NOTIFICATION_ID = 1001;

    private OFAAgent agent;

    @Override
    public void onCreate() {
        super.onCreate();
        createNotificationChannel();
        initAgent();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        // 启动前台服务
        startForeground(NOTIFICATION_ID, createNotification());

        if (intent != null) {
            String action = intent.getAction();
            if ("CONNECT".equals(action)) {
                connectAgent();
            } else if ("DISCONNECT".equals(action)) {
                disconnectAgent();
            } else if ("SET_OFFLINE".equals(action)) {
                boolean offline = intent.getBooleanExtra("offline", true);
                setOfflineMode(offline);
            }
        }

        return START_STICKY;
    }

    @Nullable
    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    @Override
    public void onDestroy() {
        if (agent != null) {
            agent.shutdown();
        }
        super.onDestroy();
    }

    private void createNotificationChannel() {
        NotificationChannel channel = new NotificationChannel(
                CHANNEL_ID,
                "OFA Agent Service",
                NotificationManager.IMPORTANCE_LOW
        );
        channel.setDescription("OFA Agent 后台服务");

        NotificationManager manager = getSystemService(NotificationManager.class);
        if (manager != null) {
            manager.createNotificationChannel(channel);
        }
    }

    private Notification createNotification() {
        return new NotificationCompat.Builder(this, CHANNEL_ID)
                .setContentTitle("OFA Agent")
                .setContentText("Agent 运行中")
                .setSmallIcon(android.R.drawable.ic_dialog_info)
                .setPriority(NotificationCompat.PRIORITY_LOW)
                .setOngoing(true)
                .build();
    }

    private void initAgent() {
        // 从 SharedPreferences 获取配置
        android.content.SharedPreferences prefs = getSharedPreferences("ofa_config", Context.MODE_PRIVATE);
        String centerAddress = prefs.getString("center_address", "192.168.1.100");
        int centerPort = prefs.getInt("center_port", 9090);
        String agentId = prefs.getString("agent_id", "android-" + android.os.Build.SERIAL);

        agent = new OFAAgent.Builder(this)
                .agentId(agentId)
                .agentName(android.os.Build.MODEL)
                .centerAddress(centerAddress)
                .centerPort(centerPort)
                .type(OFAAgent.AgentType.MOBILE)
                .offlineLevel(OfflineLevel.L4)
                .enableTools(true)
                .build();

        // 注册离线技能
        OfflineManager offlineManager = agent.getOfflineManager();
        if (offlineManager != null) {
            OfflineSkills.registerAll(offlineManager);
        }

        // 设置连接监听器
        agent.setConnectionListener(new OFAAgent.ConnectionListener() {
            @Override
            public void onConnected() {
                Log.i(TAG, "Agent connected");
                updateNotification("已连接到 Center");
            }

            @Override
            public void onDisconnected() {
                Log.w(TAG, "Agent disconnected");
                updateNotification("已断开连接");
            }

            @Override
            public void onError(String message) {
                Log.e(TAG, "Agent error: " + message);
                updateNotification("错误: " + message);
            }
        });
    }

    private void connectAgent() {
        if (agent != null && !agent.isConnected()) {
            agent.connect();
        }
    }

    private void disconnectAgent() {
        if (agent != null && agent.isConnected()) {
            agent.disconnect();
        }
    }

    private void setOfflineMode(boolean offline) {
        if (agent != null) {
            agent.setOfflineMode(offline);
            updateNotification(offline ? "离线模式" : "在线模式");
        }
    }

    private void updateNotification(String status) {
        Notification notification = new NotificationCompat.Builder(this, CHANNEL_ID)
                .setContentTitle("OFA Agent")
                .setContentText(status)
                .setSmallIcon(android.R.drawable.ic_dialog_info)
                .setPriority(NotificationCompat.PRIORITY_LOW)
                .setOngoing(true)
                .build();

        NotificationManager manager = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
        if (manager != null) {
            manager.notify(NOTIFICATION_ID, notification);
        }
    }

    /**
     * 获取 Agent 实例
     */
    public OFAAgent getAgent() {
        return agent;
    }
}