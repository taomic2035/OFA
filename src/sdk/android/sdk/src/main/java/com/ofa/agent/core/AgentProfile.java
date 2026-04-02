package com.ofa.agent.core;

import android.os.Build;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

/**
 * Agent Profile - describes the agent's identity, capabilities, and status.
 *
 * Used for:
 * - Registration with Center
 * - Capability broadcasting
 * - Agent discovery
 * - Task matching
 */
public class AgentProfile {

    // Identity
    private final String agentId;
    private final String name;
    private final AgentType type;

    // Device info
    private final DeviceInfo deviceInfo;

    // Capabilities
    private final List<Capability> capabilities;
    private final Map<String, String> skills;

    // Status
    private AgentStatus status;
    private final Map<String, Object> metadata;

    // Configuration
    private RunMode preferredRunMode;
    private boolean allowRemoteControl;
    private boolean allowPeerCommunication;

    /**
     * Agent type enum
     */
    public enum AgentType {
        FULL(1, "Full Agent", "Complete agent with all capabilities"),
        MOBILE(2, "Mobile Agent", "Android/iOS mobile device"),
        LITE(3, "Lite Agent", "Watch/band wearable device"),
        IOT(4, "IoT Agent", "Smart home IoT device"),
        EDGE(5, "Edge Agent", "Edge computing device");

        private final int value;
        private final String displayName;
        private final String description;

        AgentType(int value, String displayName, String description) {
            this.value = value;
            this.displayName = displayName;
            this.description = description;
        }

        public int getValue() { return value; }
        public String getDisplayName() { return displayName; }
        public String getDescription() { return description; }
    }

    /**
     * Run mode enum
     */
    public enum RunMode {
        STANDALONE("standalone", "完全独立运行，不连接 Center"),
        CONNECTED("connected", "连接 Center，接收远程任务"),
        HYBRID("hybrid", "混合模式，本地优先，云端增强");

        private final String value;
        private final String description;

        RunMode(String value, String description) {
            this.value = value;
            this.description = description;
        }

        public String getValue() { return value; }
        public String getDescription() { return description; }
    }

    /**
     * Agent status enum
     */
    public enum AgentStatus {
        OFFLINE(0, "离线"),
        ONLINE(1, "在线"),
        BUSY(2, "忙碌"),
        IDLE(3, "空闲"),
        MAINTENANCE(4, "维护中");

        private final int value;
        private final String displayName;

        AgentStatus(int value, String displayName) {
            this.value = value;
            this.displayName = displayName;
        }

        public int getValue() { return value; }
        public String getDisplayName() { return displayName; }
    }

    /**
     * Device information
     */
    public static class DeviceInfo {
        public final String os;
        public final String osVersion;
        public final String model;
        public final String manufacturer;
        public final String arch;
        public final int sdkVersion;
        public final long totalMemory;
        public final long availableStorage;

        public DeviceInfo(String os, String osVersion, String model, String manufacturer,
                          String arch, int sdkVersion, long totalMemory, long availableStorage) {
            this.os = os;
            this.osVersion = osVersion;
            this.model = model;
            this.manufacturer = manufacturer;
            this.arch = arch;
            this.sdkVersion = sdkVersion;
            this.totalMemory = totalMemory;
            this.availableStorage = availableStorage;
        }

        public static DeviceInfo fromCurrentDevice() {
            Runtime runtime = Runtime.getRuntime();
            return new DeviceInfo(
                "android",
                Build.VERSION.RELEASE,
                Build.MODEL,
                Build.MANUFACTURER,
                Build.SUPPORTED_ABIS[0],
                Build.VERSION.SDK_INT,
                runtime.maxMemory(),
                -1 // Would need StatFs for storage
            );
        }
    }

    /**
     * Capability definition
     */
    public static class Capability {
        public final String id;
        public final String name;
        public final String category;
        public final int priority; // 1-10, higher is more capable
        public final boolean requiresNetwork;
        public final boolean requiresLocal;

        public Capability(String id, String name, String category, int priority,
                          boolean requiresNetwork, boolean requiresLocal) {
            this.id = id;
            this.name = name;
            this.category = category;
            this.priority = priority;
            this.requiresNetwork = requiresNetwork;
            this.requiresLocal = requiresLocal;
        }

        // Common capabilities
        public static Capability UI_AUTOMATION = new Capability(
            "ui_automation", "UI Automation", "automation", 8, false, true);

        public static Capability SOCIAL_NOTIFICATION = new Capability(
            "social_notification", "Social Notification", "communication", 7, true, false);

        public static Capability LOCAL_LLM = new Capability(
            "local_llm", "Local LLM", "ai", 6, false, true);

        public static Capability CLOUD_LLM = new Capability(
            "cloud_llm", "Cloud LLM", "ai", 9, true, false);

        public static Capability INTENT_UNDERSTANDING = new Capability(
            "intent_understanding", "Intent Understanding", "ai", 7, false, true);

        public static Capability MEMORY_SYSTEM = new Capability(
            "memory_system", "Memory System", "data", 5, false, true);

        public static Capability SKILL_ORCHESTRATION = new Capability(
            "skill_orchestration", "Skill Orchestration", "automation", 8, false, true);

        public static Capability CONTACT_ACCESS = new Capability(
            "contact_access", "Contact Access", "data", 4, false, true);
    }

    /**
     * Builder for AgentProfile
     */
    public static class Builder {
        private String agentId;
        private String name;
        private AgentType type = AgentType.MOBILE;
        private DeviceInfo deviceInfo;
        private final List<Capability> capabilities = new ArrayList<>();
        private final Map<String, String> skills = new HashMap<>();
        private AgentStatus status = AgentStatus.OFFLINE;
        private final Map<String, Object> metadata = new HashMap<>();
        private RunMode preferredRunMode = RunMode.HYBRID;
        private boolean allowRemoteControl = true;
        private boolean allowPeerCommunication = true;

        public Builder agentId(String agentId) {
            this.agentId = agentId;
            return this;
        }

        public Builder name(String name) {
            this.name = name;
            return this;
        }

        public Builder type(AgentType type) {
            this.type = type;
            return this;
        }

        public Builder deviceInfo(DeviceInfo deviceInfo) {
            this.deviceInfo = deviceInfo;
            return this;
        }

        public Builder addCapability(Capability capability) {
            this.capabilities.add(capability);
            return this;
        }

        public Builder addCapabilities(List<Capability> capabilities) {
            this.capabilities.addAll(capabilities);
            return this;
        }

        public Builder addSkill(String skillId, String description) {
            this.skills.put(skillId, description);
            return this;
        }

        public Builder status(AgentStatus status) {
            this.status = status;
            return this;
        }

        public Builder metadata(String key, Object value) {
            this.metadata.put(key, value);
            return this;
        }

        public Builder preferredRunMode(RunMode mode) {
            this.preferredRunMode = mode;
            return this;
        }

        public Builder allowRemoteControl(boolean allow) {
            this.allowRemoteControl = allow;
            return this;
        }

        public Builder allowPeerCommunication(boolean allow) {
            this.allowPeerCommunication = allow;
            return this;
        }

        public AgentProfile build() {
            if (agentId == null) {
                agentId = java.util.UUID.randomUUID().toString();
            }
            if (name == null) {
                name = Build.MODEL;
            }
            if (deviceInfo == null) {
                deviceInfo = DeviceInfo.fromCurrentDevice();
            }
            return new AgentProfile(this);
        }
    }

    private AgentProfile(Builder builder) {
        this.agentId = builder.agentId;
        this.name = builder.name;
        this.type = builder.type;
        this.deviceInfo = builder.deviceInfo;
        this.capabilities = new ArrayList<>(builder.capabilities);
        this.skills = new HashMap<>(builder.skills);
        this.status = builder.status;
        this.metadata = new HashMap<>(builder.metadata);
        this.preferredRunMode = builder.preferredRunMode;
        this.allowRemoteControl = builder.allowRemoteControl;
        this.allowPeerCommunication = builder.allowPeerCommunication;
    }

    // Getters

    @NonNull
    public String getAgentId() { return agentId; }

    @NonNull
    public String getName() { return name; }

    @NonNull
    public AgentType getType() { return type; }

    @NonNull
    public DeviceInfo getDeviceInfo() { return deviceInfo; }

    @NonNull
    public List<Capability> getCapabilities() { return new ArrayList<>(capabilities); }

    @NonNull
    public Map<String, String> getSkills() { return new HashMap<>(skills); }

    @NonNull
    public AgentStatus getStatus() { return status; }

    public void setStatus(AgentStatus status) { this.status = status; }

    @NonNull
    public Map<String, Object> getMetadata() { return new HashMap<>(metadata); }

    @NonNull
    public RunMode getPreferredRunMode() { return preferredRunMode; }

    public void setPreferredRunMode(RunMode mode) { this.preferredRunMode = mode; }

    public boolean isAllowRemoteControl() { return allowRemoteControl; }

    public boolean isAllowPeerCommunication() { return allowPeerCommunication; }

    // Utility methods

    public boolean hasCapability(String capabilityId) {
        for (Capability cap : capabilities) {
            if (cap.id.equals(capabilityId)) return true;
        }
        return false;
    }

    public boolean hasSkill(String skillId) {
        return skills.containsKey(skillId);
    }

    public boolean canExecuteOffline(String capabilityId) {
        for (Capability cap : capabilities) {
            if (cap.id.equals(capabilityId) && !cap.requiresNetwork) {
                return true;
            }
        }
        return false;
    }

    /**
     * Convert to JSON for transmission
     */
    @NonNull
    public String toJson() {
        StringBuilder sb = new StringBuilder();
        sb.append("{");
        sb.append("\"agentId\":\"").append(agentId).append("\",");
        sb.append("\"name\":\"").append(name).append("\",");
        sb.append("\"type\":\"").append(type.getValue()).append("\",");
        sb.append("\"status\":\"").append(status.getValue()).append("\",");
        sb.append("\"runMode\":\"").append(preferredRunMode.getValue()).append("\",");

        sb.append("\"device\":{");
        sb.append("\"os\":\"").append(deviceInfo.os).append("\",");
        sb.append("\"model\":\"").append(deviceInfo.model).append("\"");
        sb.append("},");

        sb.append("\"capabilities\":[");
        for (int i = 0; i < capabilities.size(); i++) {
            if (i > 0) sb.append(",");
            sb.append("\"").append(capabilities.get(i).id).append("\"");
        }
        sb.append("]");

        sb.append("}");
        return sb.toString();
    }

    @NonNull
    @Override
    public String toString() {
        return String.format("AgentProfile{id=%s, name=%s, type=%s, mode=%s, status=%s, capabilities=%d}",
            agentId, name, type, preferredRunMode, status, capabilities.size());
    }
}