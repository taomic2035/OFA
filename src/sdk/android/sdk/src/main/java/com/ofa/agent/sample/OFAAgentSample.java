package com.ofa.agent.sample;

import android.app.Application;
import android.util.Log;

import com.ofa.agent.OFAAgent;
import com.ofa.agent.offline.OfflineLevel;
import com.ofa.agent.offline.OfflineManager;
import com.ofa.agent.skill.builtin.OfflineSkills;
import com.ofa.agent.tool.ToolResult;

import org.json.JSONObject;

import java.util.HashMap;
import java.util.Map;

/**
 * OFA Agent SDK 使用示例
 *
 * 这个类展示了如何在 Android 应用中集成和使用 OFA Agent SDK
 */
public class OFAAgentSample extends Application {

    private static final String TAG = "OFASample";
    private OFAAgent agent;

    @Override
    public void onCreate() {
        super.onCreate();
        initOFAAgent();
    }

    /**
     * 初始化 OFA Agent
     */
    private void initOFAAgent() {
        // 创建 Agent 实例
        agent = new OFAAgent.Builder(this)
                .agentId("my-android-device-001")
                .agentName("My Android Phone")
                .centerAddress("192.168.1.100")  // OFA Center 地址
                .centerPort(9090)
                .type(OFAAgent.AgentType.MOBILE)
                .offlineLevel(OfflineLevel.L4)   // 设置离线等级
                .enableTools(true)                // 启用 MCP/Tool 支持
                .build();

        // 注册内置离线技能
        OfflineManager offlineManager = agent.getOfflineManager();
        if (offlineManager != null) {
            OfflineSkills.registerAll(offlineManager);
            Log.i(TAG, "Offline skills registered");
        }

        // 设置连接监听器
        agent.setConnectionListener(new OFAAgent.ConnectionListener() {
            @Override
            public void onConnected() {
                Log.i(TAG, "Connected to OFA Center");
                onAgentConnected();
            }

            @Override
            public void onDisconnected() {
                Log.w(TAG, "Disconnected from OFA Center");
            }

            @Override
            public void onError(String message) {
                Log.e(TAG, "Connection error: " + message);
            }
        });

        // 连接到 Center
        agent.connect();
    }

    /**
     * Agent 连接成功后的操作
     */
    private void onAgentConnected() {
        // 获取可用工具列表
        Log.i(TAG, "Available tools: " + agent.getAvailableTools().size());

        // 示例：调用电池状态工具
        checkBatteryStatus();

        // 示例：本地离线计算
        performOfflineCalculation();
    }

    /**
     * 示例：调用电池状态工具
     */
    private void checkBatteryStatus() {
        ToolResult result = agent.callTool("battery.status", new HashMap<>());

        if (result.isSuccess()) {
            JSONObject output = result.getOutput();
            int level = output.optInt("level", -1);
            String status = output.optString("status", "unknown");

            Log.i(TAG, "Battery: " + level + "%, Status: " + status);
        } else {
            Log.e(TAG, "Battery check failed: " + result.getError());
        }
    }

    /**
     * 示例：执行离线计算
     */
    private void performOfflineCalculation() {
        OfflineManager offlineManager = agent.getOfflineManager();
        if (offlineManager == null) return;

        // 本地执行计算器技能
        String taskId = offlineManager.executeLocal("calculator", "100 * 7".getBytes());

        // 检查任务状态
        new android.os.Handler().postDelayed(() -> {
            var task = offlineManager.getTask(taskId);
            if (task != null && task.getStatus() == com.ofa.agent.offline.TaskStatus.COMPLETED) {
                String output = new String(task.getOutput());
                Log.i(TAG, "Calculation result: " + output);
            }
        }, 1000);
    }

    /**
     * 示例：使用 AI Agent 接口
     */
    public void useAIInterface() {
        var aiInterface = agent.getAIAgentInterface();
        if (aiInterface == null) return;

        // 获取工具列表 (OpenAI 函数格式)
        org.json.JSONArray functions = aiInterface.getToolsAsFunctions();
        Log.i(TAG, "Tools as functions: " + functions.length() + " available");

        // 调用工具
        Map<String, Object> args = new HashMap<>();
        args.put("text", "hello world");
        args.put("operation", "uppercase");

        ToolResult result = aiInterface.callTool("text.process", args);
        if (result.isSuccess()) {
            Log.i(TAG, "Text process result: " + result.getOutput());
        }
    }

    /**
     * 示例：处理离线模式切换
     */
    public void handleOfflineMode() {
        OfflineManager offlineManager = agent.getOfflineManager();
        if (offlineManager == null) return;

        // 添加离线模式监听器
        offlineManager.addOfflineModeListener(offline -> {
            if (offline) {
                Log.i(TAG, "Entered offline mode - using cached data and local skills");
            } else {
                Log.i(TAG, "Back online - syncing data");
                offlineManager.syncNow();
            }
        });

        // 手动设置离线模式
        // agent.setOfflineMode(true);
    }

    /**
     * 示例：使用 P2P 通信
     */
    public void useP2PCommunication() {
        var p2pClient = agent.getOfflineManager(); // P2P 通过 OfflineManager 管理

        // 发现附近设备
        // p2pClient.discoverPeers();

        // 发送消息到其他 Agent
        // p2pClient.sendMessage("agent-002", "Hello from Android".getBytes());
    }

    /**
     * 示例：约束检查
     */
    public void checkConstraints() {
        // 在执行敏感操作前检查约束
        String operation = "camera.capture";
        Map<String, Object> params = new HashMap<>();
        params.put("cameraId", "0");

        // 约束检查器会验证操作是否允许
        // 如果在离线模式下，某些操作可能被限制
        ToolResult result = agent.callTool(operation, params);

        if (!result.isSuccess()) {
            Log.w(TAG, "Operation blocked: " + result.getError());
        }
    }

    /**
     * 清理资源
     */
    @Override
    public void onTerminate() {
        if (agent != null) {
            agent.shutdown();
        }
        super.onTerminate();
    }

    /**
     * 获取 Agent 实例（供其他组件使用）
     */
    public OFAAgent getAgent() {
        return agent;
    }
}