package com.ofa.agent.grpc;

import androidx.annotation.NonNull;
import androidx.annotation.Nullable;

import java.util.List;

/**
 * Protobuf generated class stub.
 * This is a placeholder - actual implementation requires protobuf generation.
 */
public final class AgentOuterClass {

    private AgentOuterClass() {}

    // PayloadCase enum for CenterMessage
    public enum PayloadCase {
        TASK,
        CONFIG,
        BROADCAST,
        PAYLOAD_NOT_SET
    }

    // AgentMessage stub
    public static final class AgentMessage {
        private final String msgId;
        private final RegisterRequest register;
        private final HeartbeatRequest heartbeat;
        private final TaskResult taskResult;

        private AgentMessage(Builder builder) {
            this.msgId = builder.msgId;
            this.register = builder.register;
            this.heartbeat = builder.heartbeat;
            this.taskResult = builder.taskResult;
        }

        public String getMsgId() { return msgId; }
        public RegisterRequest getRegister() { return register; }
        public HeartbeatRequest getHeartbeat() { return heartbeat; }
        public TaskResult getTaskResult() { return taskResult; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String msgId = "";
            private RegisterRequest register;
            private HeartbeatRequest heartbeat;
            private TaskResult taskResult;

            public Builder setMsgId(String id) { this.msgId = id; return this; }
            public Builder setRegister(RegisterRequest r) { this.register = r; return this; }
            public Builder setHeartbeat(HeartbeatRequest h) { this.heartbeat = h; return this; }
            public Builder setTaskResult(TaskResult t) { this.taskResult = t; return this; }
            public AgentMessage build() { return new AgentMessage(this); }
        }
    }

    // CenterMessage stub
    public static final class CenterMessage {
        private final TaskAssignment task;
        private final ConfigUpdate config;
        private final BroadcastMessage broadcast;
        private final PayloadCase payloadCase;

        private CenterMessage(Builder builder) {
            this.task = builder.task;
            this.config = builder.config;
            this.broadcast = builder.broadcast;
            if (task != null) {
                this.payloadCase = PayloadCase.TASK;
            } else if (config != null) {
                this.payloadCase = PayloadCase.CONFIG;
            } else if (broadcast != null) {
                this.payloadCase = PayloadCase.BROADCAST;
            } else {
                this.payloadCase = PayloadCase.PAYLOAD_NOT_SET;
            }
        }

        public TaskAssignment getTask() { return task; }
        public ConfigUpdate getConfig() { return config; }
        public BroadcastMessage getBroadcast() { return broadcast; }
        public PayloadCase getPayloadCase() { return payloadCase; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private TaskAssignment task;
            private ConfigUpdate config;
            private BroadcastMessage broadcast;

            public Builder setTask(TaskAssignment t) { this.task = t; return this; }
            public Builder setConfig(ConfigUpdate c) { this.config = c; return this; }
            public Builder setBroadcast(BroadcastMessage b) { this.broadcast = b; return this; }
            public CenterMessage build() { return new CenterMessage(this); }
        }
    }

    // RegisterRequest stub
    public static final class RegisterRequest {
        private final String agentId;
        private final String name;
        private final int typeValue;
        private final DeviceInfo deviceInfo;
        private final List<Capability> capabilities;

        private RegisterRequest(Builder builder) {
            this.agentId = builder.agentId;
            this.name = builder.name;
            this.typeValue = builder.typeValue;
            this.deviceInfo = builder.deviceInfo;
            this.capabilities = builder.capabilities;
        }

        public String getAgentId() { return agentId; }
        public String getName() { return name; }
        public int getTypeValue() { return typeValue; }
        public DeviceInfo getDeviceInfo() { return deviceInfo; }
        public List<Capability> getCapabilitiesList() { return capabilities; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String agentId = "";
            private String name = "";
            private int typeValue = 0;
            private DeviceInfo deviceInfo;
            private List<Capability> capabilities = new java.util.ArrayList<>();

            public Builder setAgentId(String id) { this.agentId = id; return this; }
            public Builder setName(String n) { this.name = n; return this; }
            public Builder setTypeValue(int v) { this.typeValue = v; return this; }
            public Builder setDeviceInfo(DeviceInfo info) { this.deviceInfo = info; return this; }
            public Builder addAllCapabilities(List<Capability> caps) { this.capabilities = caps; return this; }
            public RegisterRequest build() { return new RegisterRequest(this); }
        }
    }

    // HeartbeatRequest stub
    public static final class HeartbeatRequest {
        private final String agentId;
        private final int statusValue;
        private final ResourceUsage resources;

        private HeartbeatRequest(Builder builder) {
            this.agentId = builder.agentId;
            this.statusValue = builder.statusValue;
            this.resources = builder.resources;
        }

        public String getAgentId() { return agentId; }
        public int getStatusValue() { return statusValue; }
        public ResourceUsage getResources() { return resources; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String agentId = "";
            private int statusValue = 0;
            private ResourceUsage resources;

            public Builder setAgentId(String id) { this.agentId = id; return this; }
            public Builder setStatusValue(int s) { this.statusValue = s; return this; }
            public Builder setResources(ResourceUsage r) { this.resources = r; return this; }
            public HeartbeatRequest build() { return new HeartbeatRequest(this); }
        }
    }

    // TaskAssignment stub
    public static final class TaskAssignment {
        private final String taskId;
        private final String skillId;
        private final com.google.protobuf.ByteString input;

        private TaskAssignment(Builder builder) {
            this.taskId = builder.taskId;
            this.skillId = builder.skillId;
            this.input = builder.input;
        }

        public String getTaskId() { return taskId; }
        public String getSkillId() { return skillId; }
        public com.google.protobuf.ByteString getInput() { return input; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String taskId = "";
            private String skillId = "";
            private com.google.protobuf.ByteString input = com.google.protobuf.ByteString.EMPTY;

            public Builder setTaskId(String id) { this.taskId = id; return this; }
            public Builder setSkillId(String id) { this.skillId = id; return this; }
            public Builder setInput(com.google.protobuf.ByteString i) { this.input = i; return this; }
            public TaskAssignment build() { return new TaskAssignment(this); }
        }
    }

    // TaskResult stub
    public static final class TaskResult {
        private final String taskId;
        private final int statusValue;
        private final String result;
        private final com.google.protobuf.ByteString output;
        private final String error;

        private TaskResult(Builder builder) {
            this.taskId = builder.taskId;
            this.statusValue = builder.statusValue;
            this.result = builder.result;
            this.output = builder.output;
            this.error = builder.error;
        }

        public String getTaskId() { return taskId; }
        public int getStatusValue() { return statusValue; }
        public String getResult() { return result; }
        public com.google.protobuf.ByteString getOutput() { return output; }
        public String getError() { return error; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String taskId = "";
            private int statusValue = 0;
            private String result = "";
            private com.google.protobuf.ByteString output = com.google.protobuf.ByteString.EMPTY;
            private String error = "";

            public Builder setTaskId(String id) { this.taskId = id; return this; }
            public Builder setStatusValue(int s) { this.statusValue = s; return this; }
            public Builder setResult(String r) { this.result = r; return this; }
            public Builder setOutput(com.google.protobuf.ByteString o) { this.output = o; return this; }
            public Builder setError(String e) { this.error = e; return this; }
            public TaskResult build() { return new TaskResult(this); }
        }
    }

    // ConfigUpdate stub
    public static final class ConfigUpdate {
        private final String configJson;

        private ConfigUpdate(Builder builder) {
            this.configJson = builder.configJson;
        }

        public String getConfigJson() { return configJson; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String configJson = "";

            public Builder setConfigJson(String json) { this.configJson = json; return this; }
            public ConfigUpdate build() { return new ConfigUpdate(this); }
        }
    }

    // BroadcastMessage stub
    public static final class BroadcastMessage {
        private final String action;
        private final String message;

        private BroadcastMessage(Builder builder) {
            this.action = builder.action;
            this.message = builder.message;
        }

        public String getAction() { return action; }
        public String getMessage() { return message; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String action = "";
            private String message = "";

            public Builder setAction(String a) { this.action = a; return this; }
            public Builder setMessage(String msg) { this.message = msg; return this; }
            public BroadcastMessage build() { return new BroadcastMessage(this); }
        }
    }

    // DeviceInfo stub
    public static final class DeviceInfo {
        private final String model;
        private final String manufacturer;
        private final String arch;
        private final String os;
        private final String osVersion;
        private final int sdkLevel;

        private DeviceInfo(Builder builder) {
            this.model = builder.model;
            this.manufacturer = builder.manufacturer;
            this.arch = builder.arch;
            this.os = builder.os;
            this.osVersion = builder.osVersion;
            this.sdkLevel = builder.sdkLevel;
        }

        public String getModel() { return model; }
        public String getManufacturer() { return manufacturer; }
        public String getArch() { return arch; }
        public String getOs() { return os; }
        public String getOsVersion() { return osVersion; }
        public int getSdkLevel() { return sdkLevel; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String model = "";
            private String manufacturer = "";
            private String arch = "";
            private String os = "";
            private String osVersion = "";
            private int sdkLevel = 0;

            public Builder setModel(String m) { this.model = m; return this; }
            public Builder setManufacturer(String m) { this.manufacturer = m; return this; }
            public Builder setArch(String a) { this.arch = a; return this; }
            public Builder setOs(String o) { this.os = o; return this; }
            public Builder setOsVersion(String v) { this.osVersion = v; return this; }
            public Builder setSdkLevel(int l) { this.sdkLevel = l; return this; }
            public DeviceInfo build() { return new DeviceInfo(this); }
        }
    }

    // ResourceUsage stub
    public static final class ResourceUsage {
        private final int cpuUsage;
        private final int memoryUsage;
        private final int batteryLevel;

        private ResourceUsage(Builder builder) {
            this.cpuUsage = builder.cpuUsage;
            this.memoryUsage = builder.memoryUsage;
            this.batteryLevel = builder.batteryLevel;
        }

        public int getCpuUsage() { return cpuUsage; }
        public int getMemoryUsage() { return memoryUsage; }
        public int getBatteryLevel() { return batteryLevel; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private int cpuUsage = 0;
            private int memoryUsage = 0;
            private int batteryLevel = 100;

            public Builder setCpuUsage(int c) { this.cpuUsage = c; return this; }
            public Builder setMemoryUsage(int m) { this.memoryUsage = m; return this; }
            public Builder setBatteryLevel(int b) { this.batteryLevel = b; return this; }
            public ResourceUsage build() { return new ResourceUsage(this); }
        }
    }

    // Capability stub
    public static final class Capability {
        private final String id;
        private final String name;
        private final String version;

        private Capability(Builder builder) {
            this.id = builder.id;
            this.name = builder.name;
            this.version = builder.version;
        }

        public String getId() { return id; }
        public String getName() { return name; }
        public String getVersion() { return version; }

        public static Builder newBuilder() { return new Builder(); }

        public static final class Builder {
            private String id = "";
            private String name = "";
            private String version = "1.0";

            public Builder setId(String i) { this.id = i; return this; }
            public Builder setName(String n) { this.name = n; return this; }
            public Builder setVersion(String v) { this.version = v; return this; }
            public Capability build() { return new Capability(this); }
        }
    }
}