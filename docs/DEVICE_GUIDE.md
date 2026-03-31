# OFA 设备接入指南

## 目录

1. [Android 设备接入](#android-设备接入)
2. [iOS 设备接入](#ios-设备接入)
3. [IoT 智能家居设备接入](#iot-智能家居设备接入)
4. [可穿戴设备接入](#可穿戴设备接入)
5. [典型应用场景](#典型应用场景)

---

## Android 设备接入

### 1. 添加依赖

在 `build.gradle` 中添加：

```gradle
dependencies {
    // OFA Agent SDK
    implementation 'com.ofa:agent-sdk:9.0.0'

    // gRPC 依赖
    implementation 'io.grpc:grpc-okhttp:1.59.0'
    implementation 'io.grpc:grpc-protobuf:1.59.0'
    implementation 'io.grpc:grpc-stub:1.59.0'

    // 协程支持
    implementation 'org.jetbrains.kotlinx:kotlinx-coroutines-android:1.7.3'
}
```

### 2. 基础接入

```java
// MainActivity.java
public class MainActivity extends AppCompatActivity {
    private OFAAgent agent;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        // 创建 Agent
        agent = new OFAAgent.Builder(this)
            .agentId("android-001")           // 可选，不填自动生成
            .agentName("我的手机")             // 可选，默认使用设备型号
            .type(OFAAgent.AgentType.MOBILE)  // Agent 类型
            .centerAddress("192.168.1.100")   // Center 服务器地址
            .centerPort(9090)                  // gRPC 端口
            .build();

        // 设置连接监听
        agent.setConnectionListener(new OFAAgent.ConnectionListener() {
            @Override
            public void onConnected() {
                Log.i("OFA", "已连接到 Center");
                runOnUiThread(() ->
                    Toast.makeText(MainActivity.this, "已连接", Toast.LENGTH_SHORT).show()
                );
            }

            @Override
            public void onDisconnected() {
                Log.i("OFA", "已断开连接");
            }

            @Override
            public void onError(String message) {
                Log.e("OFA", "连接错误: " + message);
            }
        });

        // 设置任务监听
        agent.setTaskListener(new OFAAgent.TaskListener() {
            @Override
            public void onTaskReceived(String taskId, String skillId) {
                Log.i("OFA", "收到任务: " + taskId);
            }

            @Override
            public void onTaskCompleted(String taskId) {
                Log.i("OFA", "任务完成: " + taskId);
            }

            @Override
            public void onTaskFailed(String taskId, String error) {
                Log.e("OFA", "任务失败: " + taskId + ", " + error);
            }
        });

        // 连接到 Center
        agent.connect();
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (agent != null) {
            agent.disconnect();
        }
    }
}
```

### 3. 注册自定义技能

```java
// 定义技能执行器
public class CameraSkill implements SkillExecutor {

    @Override
    public String getSkillId() {
        return "device.camera";
    }

    @Override
    public String[] getActions() {
        return new String[]{"take_photo", "start_record", "stop_record"};
    }

    @Override
    public byte[] execute(byte[] input) throws SkillExecutionException {
        try {
            // 解析输入
            JSONObject params = new JSONObject(new String(input));
            String action = params.getString("action");

            switch (action) {
                case "take_photo":
                    return takePhoto();
                case "start_record":
                    return startRecording();
                case "stop_record":
                    return stopRecording();
                default:
                    throw new SkillExecutionException("未知操作: " + action);
            }
        } catch (Exception e) {
            throw new SkillExecutionException("执行失败", e);
        }
    }

    private byte[] takePhoto() {
        // 实现拍照逻辑
        JSONObject result = new JSONObject();
        try {
            result.put("success", true);
            result.put("photo_path", "/sdcard/DCIM/photo.jpg");
        } catch (JSONException e) {
            e.printStackTrace();
        }
        return result.toString().getBytes();
    }

    private byte[] startRecording() {
        // 实现开始录像逻辑
        return "{\"recording\": true}".getBytes();
    }

    private byte[] stopRecording() {
        // 实现停止录像逻辑
        return "{\"recording\": false, \"video_path\": \"/sdcard/DCIM/video.mp4\"}".getBytes();
    }
}

// 注册技能
agent.registerSkill("device.camera", new CameraSkill());
```

### 4. 后台服务

```java
// AgentService.java - 后台保活
public class AgentService extends Service {
    private OFAAgent agent;

    @Override
    public void onCreate() {
        super.onCreate();

        // 初始化 Agent
        agent = new OFAAgent.Builder(this)
            .agentId("android-service-001")
            .centerAddress("192.168.1.100")
            .centerPort(9090)
            .build();

        // 注册技能
        agent.registerSkill("device.location", new LocationSkill(this));
        agent.registerSkill("device.notification", new NotificationSkill(this));

        agent.connect();
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        // 前台服务通知
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel channel = new NotificationChannel(
                "ofa_agent", "OFA Agent", NotificationManager.IMPORTANCE_LOW
            );
            NotificationManager manager = getSystemService(NotificationManager.class);
            manager.createNotificationChannel(channel);

            Notification notification = new Notification.Builder(this, "ofa_agent")
                .setContentTitle("OFA Agent 运行中")
                .setContentText("已连接到 Center")
                .setSmallIcon(R.drawable.ic_agent)
                .build();

            startForeground(1, notification);
        }

        return START_STICKY;
    }

    @Override
    public void onDestroy() {
        super.onDestroy();
        if (agent != null) {
            agent.disconnect();
        }
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }
}

// AndroidManifest.xml
<service android:name=".AgentService" android:foregroundServiceType="location" />
```

### 5. Kotlin 示例

```kotlin
// 使用 Kotlin 协程
class MainActivity : AppCompatActivity() {
    private var agent: OFAAgent? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        agent = OFAAgent.Builder(this)
            .agentId("kotlin-001")
            .centerAddress("192.168.1.100")
            .centerPort(9090)
            .build()
            .apply {
                setConnectionListener(object : OFAAgent.ConnectionListener {
                    override fun onConnected() {
                        lifecycleScope.launch {
                            Toast.makeText(this@MainActivity, "已连接", Toast.LENGTH_SHORT).show()
                        }
                    }

                    override fun onDisconnected() {
                        Log.w("OFA", "连接断开")
                    }

                    override fun onError(message: String) {
                        Log.e("OFA", "错误: $message")
                    }
                })

                // 注册技能
                registerSkill("device.sensor", SensorSkill(this@MainActivity))
                connect()
            }
    }
}

// Kotlin 技能实现
class SensorSkill(private val context: Context) : SkillExecutor {
    private val sensorManager = context.getSystemService(Context.SENSOR_SERVICE) as SensorManager

    override fun getSkillId() = "device.sensor"

    override fun execute(input: ByteArray): ByteArray {
        val params = JSONObject(String(input))
        val sensorType = params.optString("sensor", "accelerometer")

        return when (sensorType) {
            "accelerometer" -> getAccelerometerData()
            "gyroscope" -> getGyroscopeData()
            "light" -> getLightData()
            else -> JSONObject().put("error", "未知传感器").toString().toByteArray()
        }
    }

    private fun getAccelerometerData(): ByteArray {
        // 返回加速度计数据
        return JSONObject()
            .put("x", 0.5)
            .put("y", 9.8)
            .put("z", 0.1)
            .toString()
            .toByteArray()
    }
}
```

---

## iOS 设备接入

### 1. 添加依赖

**Package.swift:**

```swift
dependencies: [
    .package(url: "https://github.com/ofa/ofa-ios-sdk.git", from: "9.0.0"),
    .package(url: "https://github.com/grpc/grpc-swift.git", from: "1.20.0"),
]
```

或使用 CocoaPods:

```ruby
pod 'OFAAgent', '~> 9.0'
pod 'GRPC', '~> 1.59'
```

### 2. 基础接入

```swift
import OFAAgent
import UIKit

class AppDelegate: UIResponder, UIApplicationDelegate {
    var agent: OFAAgent?

    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {

        // 创建 Agent
        agent = OFAAgent(
            agentId: "ios-001",
            name: UIDevice.current.name,
            type: .mobile,
            centerAddress: "192.168.1.100",
            centerPort: 9090
        )

        // 设置代理
        agent?.delegate = self
        agent?.taskDelegate = self

        // 连接
        Task {
            do {
                try await agent?.connect()
                print("已连接到 Center")
            } catch {
                print("连接失败: \(error)")
            }
        }

        return true
    }
}

// MARK: - OFAAgentDelegate
extension AppDelegate: OFAAgentDelegate {
    nonisolated func agent(_ agent: OFAAgent, didChangeConnectionState state: OFAAgent.ConnectionState) {
        Task { @MainActor in
            switch state {
            case .connected:
                print("已连接")
            case .disconnected:
                print("已断开")
            case .connecting:
                print("连接中...")
            case .reconnecting:
                print("重连中...")
            case .error(let error):
                print("错误: \(error.localizedDescription)")
            }
        }
    }
}

// MARK: - OFAAgentTaskDelegate
extension AppDelegate: OFAAgentTaskDelegate {
    nonisolated func agent(_ agent: OFAAgent, didReceiveTask taskId: String, skillId: String) {
        print("收到任务: \(taskId), 技能: \(skillId)")
    }

    nonisolated func agent(_ agent: OFAAgent, didCompleteTask taskId: String) {
        print("任务完成: \(taskId)")
    }

    nonisolated func agent(_ agent: OFAAgent, didFailTask taskId: String, error: Error) {
        print("任务失败: \(taskId), \(error)")
    }
}
```

### 3. 注册自定义技能

```swift
// 定义技能协议实现
struct CameraSkill: SkillExecutor {
    let skillId = "device.camera"
    let skillName = "相机控制"
    let category = "hardware"

    func execute(_ input: Data) async throws -> Data {
        let params = try JSONDecoder().decode(CameraParams.self, from: input)

        switch params.action {
        case "takePhoto":
            return try await takePhoto()
        case "switchCamera":
            return try await switchCamera()
        default:
            throw OFAError.executionFailed("未知操作: \(params.action)")
        }
    }

    private func takePhoto() async throws -> Data {
        // 使用 AVCaptureSession 拍照
        let result = CameraResult(
            success: true,
            imagePath: "photo_\(Date().timeIntervalSince1970).jpg"
        )
        return try JSONEncoder().encode(result)
    }

    private func switchCamera() async throws -> Data {
        return try JSONEncoder().encode(["switched": true])
    }
}

struct CameraParams: Codable {
    let action: String
}

struct CameraResult: Codable {
    let success: Bool
    let imagePath: String
}

// 注册技能
agent?.registerSkill(CameraSkill())
agent?.registerSkill(LocationSkill())
agent?.registerSkill(HealthSkill())
```

### 4. 后台模式

```swift
// 配置后台模式
// 1. 在 Info.plist 添加:
/*
 <key>UIBackgroundModes</key>
 <array>
     <string>fetch</string>
     <string>processing</string>
     <string>location</string>
 </array>
 */

// 2. 配置后台任务
class AppDelegate: UIResponder, UIApplicationDelegate {
    func application(_ application: UIApplication,
                     didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?) -> Bool {

        // 注册后台任务
        BGTaskScheduler.shared.register(forTaskWithIdentifier: "com.ofa.agent.heartbeat",
                                        using: nil) { task in
            self.handleHeartbeat(task: task as! BGAppRefreshTask)
        }

        return true
    }

    private func handleHeartbeat(task: BGAppRefreshTask) {
        // 发送心跳
        Task {
            try? await agent?.sendHeartbeat()
            task.setTaskCompleted(success: true)
        }
    }

    func scheduleBackgroundTask() {
        let request = BGAppRefreshTaskRequest(identifier: "com.ofa.agent.heartbeat")
        request.earliestBeginDate = Date(timeIntervalSinceNow: 30 * 60) // 30分钟后

        do {
            try BGTaskScheduler.shared.submit(request)
        } catch {
            print("无法调度后台任务: \(error)")
        }
    }
}
```

### 5. SwiftUI 集成

```swift
import SwiftUI
import OFAAgent

struct ContentView: View {
    @StateObject private var agentManager = AgentManager()

    var body: some View {
        VStack(spacing: 20) {
            Image(systemName: agentManager.isConnected ? "wifi" : "wifi.slash")
                .font(.system(size: 60))
                .foregroundColor(agentManager.isConnected ? .green : .red)

            Text(agentManager.statusText)
                .font(.headline)

            Button(action: {
                Task {
                    if agentManager.isConnected {
                        await agentManager.disconnect()
                    } else {
                        await agentManager.connect()
                    }
                }
            }) {
                Text(agentManager.isConnected ? "断开连接" : "连接")
                    .padding()
                    .background(agentManager.isConnected ? Color.red : Color.green)
                    .foregroundColor(.white)
                    .cornerRadius(10)
            }

            List(agentManager.registeredSkills, id: \.self) { skill in
                HStack {
                    Image(systemName: "gearshape")
                    Text(skill)
                }
            }
        }
        .padding()
    }
}

class AgentManager: ObservableObject {
    @Published var isConnected = false
    @Published var statusText = "未连接"
    @Published var registeredSkills: [String] = []

    private var agent: OFAAgent?

    init() {
        agent = OFAAgent(
            agentId: "swiftui-001",
            centerAddress: "192.168.1.100",
            centerPort: 9090
        )

        agent?.delegate = AgentDelegate { [weak self] state in
            DispatchQueue.main.async {
                self?.isConnected = state == .connected
                self?.statusText = self?.isConnected == true ? "已连接" : "已断开"
            }
        }

        // 注册技能
        agent?.registerSkill(CameraSkill())
        agent?.registerSkill(LocationSkill())
        registeredSkills = agent?.getRegisteredSkills() ?? []
    }

    func connect() async {
        try? await agent?.connect()
    }

    func disconnect() async {
        agent?.disconnect()
    }
}

struct AgentDelegate: OFAAgentDelegate {
    let onStateChange: (OFAAgent.ConnectionState) -> Void

    nonisolated func agent(_ agent: OFAAgent, didChangeConnectionState state: OFAAgent.ConnectionState) {
        onStateChange(state)
    }
}
```

---

## IoT 智能家居设备接入

### 1. 设备类型

OFA IoT SDK 支持多种智能家居设备：

| 设备类型 | 类型ID | 能力 |
|---------|--------|------|
| 智能灯 | `light` | 开关、亮度、颜色 |
| 智能插座 | `socket` | 开关、电量统计 |
| 温度传感器 | `temperature_sensor` | 温度上报 |
| 智能门锁 | `door_lock` | 开锁、状态查询 |
| 温控器 | `thermostat` | 温度设置、模式切换 |
| 窗帘 | `curtain` | 开合控制 |
| 摄像头 | `camera` | 视频、截图 |
| 烟雾报警器 | `smoke_detector` | 报警、状态 |

### 2. 基础接入

```go
package main

import (
    "context"
    "fmt"
    "time"
    "ofa/src/sdk/iot"
)

func main() {
    // 创建 IoT Agent
    config := iot.IoTConfig{
        DeviceID:      "smart-light-001",
        DeviceType:    "light",
        DeviceName:    "客厅主灯",
        CenterURL:     "http://192.168.1.100:8080",
        MQTTBroker:    "tcp://192.168.1.100:1883",
        MQTTClientID:  "smart-light-001",
        MQTTUser:      "ofa",
        MQTTPassword:  "password",
        MQTTQoS:       1,
        KeepAlive:     60 * time.Second,
        ShadowEnabled: true,
        ShadowSyncInt: 30 * time.Second,
    }

    agent := iot.NewIoTAgent(config)

    // 设置 MQTT 客户端
    mqttClient := iot.NewPahoMQTTClient(config)
    agent.SetMQTTClient(mqttClient)

    // 启动 Agent
    ctx := context.Background()
    if err := agent.Start(ctx); err != nil {
        panic(err)
    }

    // 设置初始属性
    agent.SetProperty("power", false)
    agent.SetProperty("brightness", 80)
    agent.SetProperty("color", "#FFFFFF")

    fmt.Println("智能灯已上线")

    // 保持运行
    select {}
}
```

### 3. 智能灯设备

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "ofa/src/sdk/iot"
    "time"
)

// SmartLight 智能灯实现
type SmartLight struct {
    agent     *iot.IoTAgent
    power     bool
    brightness int
    color     string
}

func NewSmartLight(deviceID, mqttBroker string) *SmartLight {
    light := &SmartLight{
        power:      false,
        brightness: 100,
        color:      "#FFFFFF",
    }

    config := iot.IoTConfig{
        DeviceID:      deviceID,
        DeviceType:    "light",
        DeviceName:    "智能灯",
        MQTTBroker:    mqttBroker,
        MQTTQoS:       1,
        ShadowEnabled: true,
    }

    light.agent = iot.NewIoTAgent(config)

    return light
}

func (l *SmartLight) Start(ctx context.Context) error {
    // 设置初始属性
    l.agent.SetProperty("power", l.power)
    l.agent.SetProperty("brightness", l.brightness)
    l.agent.SetProperty("color", l.color)
    l.agent.SetProperty("online", true)

    // 启动 Agent
    return l.agent.Start(ctx)
}

func (l *SmartLight) SetPower(on bool) error {
    l.power = on
    l.agent.SetProperty("power", on)

    // 发布状态变更事件
    l.agent.PublishEvent("power_change", map[string]interface{}{
        "power": on,
        "time":  time.Now().Unix(),
    })

    // 实际控制硬件
    fmt.Printf("灯光 %s\n", map[bool]string{true: "开启", false: "关闭"}[on])

    return nil
}

func (l *SmartLight) SetBrightness(level int) error {
    if level < 0 || level > 100 {
        return fmt.Errorf("亮度范围: 0-100")
    }

    l.brightness = level
    l.agent.SetProperty("brightness", level)

    // 实际控制硬件
    fmt.Printf("亮度设置为: %d%%\n", level)

    return nil
}

func (l *SmartLight) SetColor(hexColor string) error {
    l.color = hexColor
    l.agent.SetProperty("color", hexColor)

    fmt.Printf("颜色设置为: %s\n", hexColor)

    return nil
}

// 使用示例
func main() {
    light := NewSmartLight("living-room-light", "tcp://192.168.1.100:1883")

    ctx := context.Background()
    light.Start(ctx)

    // 控制
    light.SetPower(true)
    light.SetBrightness(80)
    light.SetColor("#FF6600")

    // 发布遥测数据
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        light.agent.PublishTelemetry(map[string]interface{}{
            "power_consumption": 12.5,
            "temperature":       35.2,
            "uptime":            time.Now().Unix(),
        })
    }
}
```

### 4. 温度传感器

```go
package main

import (
    "context"
    "fmt"
    "math/rand"
    "ofa/src/sdk/iot"
    "time"
)

// TemperatureSensor 温度传感器
type TemperatureSensor struct {
    agent      *iot.IoTAgent
    deviceID   string
    temp       float64
    humidity   float64
}

func NewTemperatureSensor(deviceID, mqttBroker string) *TemperatureSensor {
    return &TemperatureSensor{
        deviceID: deviceID,
        temp:     25.0,
        humidity: 60.0,
    }
}

func (s *TemperatureSensor) Start(ctx context.Context) error {
    config := iot.IoTConfig{
        DeviceID:      s.deviceID,
        DeviceType:    "temperature_sensor",
        DeviceName:    "温湿度传感器",
        MQTTBroker:    mqttBroker,
        ShadowEnabled: true,
    }

    s.agent = iot.NewIoTAgent(config)

    // 设置属性
    s.agent.SetProperty("temperature", s.temp)
    s.agent.SetProperty("humidity", s.humidity)
    s.agent.SetProperty("unit", "celsius")

    if err := s.agent.Start(ctx); err != nil {
        return err
    }

    // 定时上报数据
    go s.sensorLoop(ctx)

    return nil
}

func (s *TemperatureSensor) sensorLoop(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // 模拟读取传感器
            s.temp = 20.0 + rand.Float64()*10
            s.humidity = 40.0 + rand.Float64()*40

            // 更新属性
            s.agent.SetProperty("temperature", s.temp)
            s.agent.SetProperty("humidity", s.humidity)

            // 发布遥测数据
            s.agent.PublishTelemetry(map[string]interface{}{
                "temperature": s.temp,
                "humidity":    s.humidity,
                "timestamp":   time.Now().Unix(),
            })

            fmt.Printf("上报: 温度=%.1f°C, 湿度=%.1f%%\n", s.temp, s.humidity)
        }
    }
}

func main() {
    sensor := NewTemperatureSensor("temp-sensor-001", "tcp://192.168.1.100:1883")

    ctx := context.Background()
    sensor.Start(ctx)

    select {}
}
```

### 5. 设备影子同步

```go
// 设备影子使用示例
func shadowExample() {
    config := iot.IoTConfig{
        DeviceID:      "device-001",
        MQTTBroker:    "tcp://192.168.1.100:1883",
        ShadowEnabled: true,
    }

    agent := iot.NewIoTAgent(config)
    agent.Start(context.Background())

    // 设置报告状态 (设备实际状态)
    agent.SetProperty("power", true)
    agent.SetProperty("mode", "auto")

    // 获取影子
    shadow := agent.GetShadow()
    fmt.Printf("报告状态: %v\n", shadow.Reported)
    fmt.Printf("期望状态: %v\n", shadow.Desired)
    fmt.Printf("差异状态: %v\n", shadow.Delta)

    // 当云端设置期望状态时，自动计算 Delta
    // 设备需要处理 Delta 并同步

    // 监听期望状态变更
    go func() {
        for {
            time.Sleep(5 * time.Second)

            desired, err := agent.GetDesiredProperty("power")
            if err == nil {
                // 检查是否需要同步
                reported, _ := agent.GetProperty("power")
                if desired != reported {
                    fmt.Printf("需要同步: 期望=%v, 实际=%v\n", desired, reported)
                    // 执行同步操作
                    agent.SetProperty("power", desired)
                }
            }
        }
    }()
}
```

---

## 可穿戴设备接入

### 1. Lite Agent 配置

```go
package main

import (
    "context"
    "fmt"
    "ofa/src/sdk/lite"
    "time"
)

func main() {
    // 创建 Lite Agent 配置
    config := lite.LiteConfig{
        ServerURL:      "192.168.1.100:9090",
        AgentID:        "watch-001",
        DeviceType:     "watch",
        HeartbeatInt:   60 * time.Second,  // 低功耗: 60秒心跳
        ReconnectInt:   30 * time.Second,
        MaxReconnect:   5,
        BatterySaver:   true,              // 开启省电模式
        CompressData:   true,              // 数据压缩
        BufferSize:     10,                // 小缓冲区
        MaxMessageSize: 4096,              // 4KB 最大消息
    }

    agent := lite.NewLiteAgent(config)

    // 注册内置技能
    agent.RegisterSkill(&lite.HeartRateSkill{})
    agent.RegisterSkill(&lite.StepCountSkill{})
    agent.RegisterSkill(&lite.LocationSkill{})
    agent.RegisterSkill(&lite.NotificationSkill{})

    // 设置连接
    conn := lite.NewTCPConnection(config.ServerURL)
    agent.SetConnection(conn)

    // 启动 Agent
    ctx := context.Background()
    if err := agent.Start(ctx); err != nil {
        panic(err)
    }

    fmt.Println("手表 Agent 已启动")

    // 模拟传感器数据更新
    go func() {
        ticker := time.NewTicker(5 * time.Second)
        for range ticker.C {
            // 更新心率
            agent.UpdateSensorData("heart_rate", map[string]interface{}{
                "bpm":   72 + rand.Intn(10),
                "status": "normal",
            })

            // 更新步数
            agent.UpdateSensorData("step_count", map[string]interface{}{
                "steps":    8500,
                "distance": 6.2,
                "calories": 320,
            })

            // 更新电池
            agent.UpdateBattery(85, false, 3.8, 25.0)
        }
    }()

    select {}
}
```

### 2. 自定义技能

```go
// 自定义健康监测技能
type HealthMonitorSkill struct{}

func (s *HealthMonitorSkill) ID() string        { return "health.monitor" }
func (s *HealthMonitorSkill) Actions() []string { return []string{"measure_all", "get_status"} }
func (s *HealthMonitorSkill) PowerConsumption() int { return 3 } // 中等功耗

func (s *HealthMonitorSkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
    switch action {
    case "measure_all":
        return map[string]interface{}{
            "heart_rate":   72,
            "blood_oxygen": 98,
            "stress":       35,
            "timestamp":    time.Now().Unix(),
        }, nil

    case "get_status":
        return map[string]interface{}{
            "health_score": 85,
            "last_check":   time.Now().Add(-1 * time.Hour).Unix(),
        }, nil

    default:
        return nil, fmt.Errorf("未知操作: %s", action)
    }
}

// 睡眠监测技能
type SleepMonitorSkill struct{}

func (s *SleepMonitorSkill) ID() string        { return "health.sleep" }
func (s *SleepMonitorSkill) Actions() []string { return []string{"start", "stop", "get_stats"} }
func (s *SleepMonitorSkill) PowerConsumption() int { return 2 } // 低功耗

func (s *SleepMonitorSkill) Execute(ctx context.Context, action string, params interface{}) (interface{}, error) {
    switch action {
    case "start":
        return map[string]interface{}{
            "sleep_monitoring": true,
            "started_at":       time.Now().Unix(),
        }, nil

    case "stop":
        return map[string]interface{}{
            "sleep_monitoring": false,
            "duration_minutes": 480,
            "deep_sleep_pct":   25,
            "light_sleep_pct":  55,
            "rem_sleep_pct":    20,
        }, nil

    case "get_stats":
        return map[string]interface{}{
            "avg_duration":    460,
            "avg_deep_sleep":  22,
            "sleep_quality":   82,
        }, nil

    default:
        return nil, fmt.Errorf("未知操作: %s", action)
    }
}
```

### 3. 省电模式

```go
// 省电模式示例
func powerSavingExample() {
    config := lite.LiteConfig{
        BatterySaver: true,
    }

    agent := lite.NewLiteAgent(config)
    agent.Start(context.Background())

    // 根据电量自动调整
    for {
        time.Sleep(30 * time.Second)

        battery := agent.GetBatteryInfo()

        switch {
        case battery.Level > 50:
            // 正常模式
            fmt.Println("正常模式: 所有功能可用")

        case battery.Level > 20:
            // 省电模式
            fmt.Println("省电模式: 限制GPS和心率监测频率")
            agent.DisableSensor("gps")

        case battery.Level > 10:
            // 超级省电
            fmt.Println("超级省电: 仅保留基本通信")
            agent.DisableSensor("heart_rate")
            agent.DisableSensor("gps")

        default:
            // 紧急模式
            fmt.Println("紧急模式: 仅心跳")
        }
    }
}
```

---

## 典型应用场景

### 场景1: 智能家居控制系统

```go
// 智能家居控制中心
package main

import (
    "context"
    "fmt"
    "ofa/pkg/llm"
    "ofa/src/sdk/iot"
    "time"
)

type SmartHomeController struct {
    devices map[string]*iot.IoTAgent
    llm     *llm.Manager
}

func NewSmartHomeController() *SmartHomeController {
    return &SmartHomeController{
        devices: make(map[string]*iot.IoTAgent),
    }
}

func (c *SmartHomeController) AddDevice(deviceID, deviceType, name string, mqttBroker string) {
    config := iot.IoTConfig{
        DeviceID:      deviceID,
        DeviceType:    deviceType,
        DeviceName:    name,
        MQTTBroker:    mqttBroker,
        ShadowEnabled: true,
    }

    agent := iot.NewIoTAgent(config)
    agent.Start(context.Background())

    c.devices[deviceID] = agent
}

func (c *SmartHomeController) ControlLight(deviceID string, on bool, brightness int, color string) error {
    agent, ok := c.devices[deviceID]
    if !ok {
        return fmt.Errorf("设备不存在: %s", deviceID)
    }

    // 设置设备影子期望状态
    agent.UpdateDesired(map[string]interface{}{
        "power":      on,
        "brightness": brightness,
        "color":      color,
    })

    return nil
}

func (c *SmartHomeController) GetDeviceStatus(deviceID string) (map[string]interface{}, error) {
    agent, ok := c.devices[deviceID]
    if !ok {
        return nil, fmt.Errorf("设备不存在: %s", deviceID)
    }

    return agent.GetAllProperties(), nil
}

// 场景联动
func (c *SmartHomeController) SceneGoodNight() {
    // 关闭所有灯
    for id, agent := range c.devices {
        if agent.GetProperties().Capabilities != nil {
            agent.UpdateDesired(map[string]interface{}{
                "power": false,
            })
            fmt.Printf("关闭设备: %s\n", id)
        }
    }

    // 设置空调
    if ac, ok := c.devices["ac-001"]; ok {
        ac.UpdateDesired(map[string]interface{}{
            "power":      true,
            "mode":       "sleep",
            "temperature": 26,
        })
    }

    // 启动安防
    if security, ok := c.devices["security-001"]; ok {
        security.UpdateDesired(map[string]interface{}{
            "armed": true,
        })
    }
}

func main() {
    controller := NewSmartHomeController()

    // 添加设备
    controller.AddDevice("light-001", "light", "客厅灯", "tcp://192.168.1.100:1883")
    controller.AddDevice("light-002", "light", "卧室灯", "tcp://192.168.1.100:1883")
    controller.AddDevice("ac-001", "thermostat", "空调", "tcp://192.168.1.100:1883")
    controller.AddDevice("lock-001", "door_lock", "门锁", "tcp://192.168.1.100:1883")

    // 控制灯光
    controller.ControlLight("light-001", true, 80, "#FFAA00")

    // 查询状态
    status, _ := controller.GetDeviceStatus("light-001")
    fmt.Printf("灯光状态: %v\n", status)

    // 执行场景
    controller.SceneGoodNight()
}
```

### 场景2: 健康监测应用

```kotlin
// Android 健康监测应用
class HealthMonitorApp : Application() {
    private var agent: OFAAgent? = null

    override fun onCreate() {
        super.onCreate()

        agent = OFAAgent.Builder(this)
            .agentId("health-001")
            .centerAddress("192.168.1.100")
            .centerPort(9090)
            .build()
            .apply {
                // 注册健康技能
                registerSkill("health.heart_rate", HeartRateSkill())
                registerSkill("health.steps", StepsSkill())
                registerSkill("health.sleep", SleepSkill())

                connect()
            }
    }
}

// 心率监测技能
class HeartRateSkill : SkillExecutor {
    private val healthConnectClient by lazy {
        HealthConnectClient.getOrCreate(context)
    }

    override fun getSkillId() = "health.heart_rate"

    override fun execute(input: ByteArray): ByteArray {
        val params = JSONObject(String(input))
        val action = params.optString("action", "get_latest")

        return when (action) {
            "get_latest" -> getLatestHeartRate()
            "get_history" -> getHistory(params.optInt("hours", 24))
            "start_monitor" -> startMonitor()
            else -> JSONObject().put("error", "Unknown action").toString().toByteArray()
        }
    }

    private fun getLatestHeartRate(): ByteArray {
        // 从 Health Connect 读取
        val request = ReadRecordsRequest(
            recordType = HeartRateRecord::class,
            timeRangeFilter = TimeRangeFilter.between(
                Instant.now().minus(1, ChronoUnit.HOURS),
                Instant.now()
            )
        )

        val response = healthConnectClient.readRecords(request)
        val latest = response.records.lastOrNull()

        return JSONObject().apply {
            put("bpm", latest?.samples?.lastOrNull()?.beatsPerMinute ?: 0)
            put("timestamp", Instant.now().toEpochMilli())
        }.toString().toByteArray()
    }
}
```

### 场景3: 工业物联网

```go
// 工业设备监控
package main

import (
    "context"
    "fmt"
    "ofa/src/sdk/iot"
    "time"
)

// IndustrialDevice 工业设备
type IndustrialDevice struct {
    agent   *iot.IoTAgent
    metrics *DeviceMetrics
}

type DeviceMetrics struct {
    Temperature   float64 `json:"temperature"`
    Pressure      float64 `json:"pressure"`
    Vibration     float64 `json:"vibration"`
    PowerConsumption float64 `json:"power_consumption"`
    Runtime       int64   `json:"runtime"`
}

func NewIndustrialDevice(deviceID, mqttBroker string) *IndustrialDevice {
    config := iot.IoTConfig{
        DeviceID:      deviceID,
        DeviceType:    "industrial_equipment",
        DeviceName:    "生产线设备",
        MQTTBroker:    mqttBroker,
        MQTTQoS:       2, // 最高 QoS 确保数据可靠
        ShadowEnabled: true,
    }

    return &IndustrialDevice{
        agent:   iot.NewIoTAgent(config),
        metrics: &DeviceMetrics{},
    }
}

func (d *IndustrialDevice) Start(ctx context.Context) error {
    d.agent.Start(ctx)

    // 定时上报指标
    go d.metricsLoop(ctx)

    // 异常检测
    go d.anomalyDetection(ctx)

    return nil
}

func (d *IndustrialDevice) metricsLoop(ctx context.Context) {
    ticker := time.NewTicker(10 * time.Second)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // 采集数据
            d.collectMetrics()

            // 更新属性
            d.agent.SetProperty("temperature", d.metrics.Temperature)
            d.agent.SetProperty("pressure", d.metrics.Pressure)
            d.agent.SetProperty("vibration", d.metrics.Vibration)
            d.agent.SetProperty("power", d.metrics.PowerConsumption)

            // 发布遥测
            d.agent.PublishTelemetry(map[string]interface{}{
                "temperature":       d.metrics.Temperature,
                "pressure":          d.metrics.Pressure,
                "vibration":         d.metrics.Vibration,
                "power_consumption": d.metrics.PowerConsumption,
                "runtime":           d.metrics.Runtime,
            })
        }
    }
}

func (d *IndustrialDevice) anomalyDetection(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            // 检测异常
            if d.metrics.Temperature > 80 {
                d.agent.PublishEvent("alert", map[string]interface{}{
                    "type":    "high_temperature",
                    "value":   d.metrics.Temperature,
                    "threshold": 80,
                    "severity": "warning",
                })
            }

            if d.metrics.Vibration > 5.0 {
                d.agent.PublishEvent("alert", map[string]interface{}{
                    "type":    "high_vibration",
                    "value":   d.metrics.Vibration,
                    "threshold": 5.0,
                    "severity": "critical",
                })
            }
        }
    }
}

func (d *IndustrialDevice) collectMetrics() {
    // 实际采集逻辑
    // d.metrics.Temperature = readTemperature()
    // ...
}
```

### 场景4: 跨设备协同控制

```go
// 手机控制智能家居
package main

import (
    "context"
    "fmt"
    "time"

    "ofa/pkg/collab"
    "ofa/src/sdk/iot"
)

func main() {
    // 手机作为控制中心
    phoneAgent := NewMobileAgent("phone-001")
    phoneAgent.Connect("192.168.1.100:9090")

    // 创建协作任务
    collabMgr := collab.NewCollaborationManager()

    // "回家模式" 场景
    homeScene := &collab.CreateCollabRequest{
        Name:        "回家模式",
        Type:        collab.CollabTypeParallel,
        Description: "打开灯光、空调、热水器",
        Tasks: []*collab.CollabTask{
            {
                ID:       "light-on",
                Name:     "开灯",
                SkillID:  "light.control",
                Operation: "turn_on",
                Input: map[string]interface{}{
                    "device_id":  "light-001",
                    "brightness": 100,
                },
            },
            {
                ID:       "ac-on",
                Name:     "开空调",
                SkillID:  "thermostat.control",
                Operation: "set_mode",
                Input: map[string]interface{}{
                    "device_id":   "ac-001",
                    "mode":        "cool",
                    "temperature": 24,
                },
            },
            {
                ID:       "heater-on",
                Name:     "开热水器",
                SkillID:  "heater.control",
                Operation: "turn_on",
                Input: map[string]interface{}{
                    "device_id":     "heater-001",
                    "target_temp":   55,
                },
            },
        },
    }

    // 创建协作
    collab, _ := collabMgr.CreateCollaboration(context.Background(), homeScene)

    // 启动执行
    collabMgr.StartCollaboration(context.Background(), collab.ID)

    // 等待完成
    time.Sleep(5 * time.Second)

    result, _ := collabMgr.GetCollaboration(collab.ID)
    fmt.Printf("场景执行结果: 成功=%v\n", result.Result.Success)
}
```

---

## 常见问题

### Q: 如何处理网络不稳定?

```go
// Android 自动重连
agent.setConnectionListener(new OFAAgent.ConnectionListener() {
    @Override
    public void onError(String message) {
        // 5秒后重连
        handler.postDelayed(() -> {
            if (!agent.isConnected()) {
                agent.connect();
            }
        }, 5000);
    }
});
```

### Q: 如何节省电量?

```go
// 可穿戴设备省电
config := lite.LiteConfig{
    BatterySaver: true,
    HeartbeatInt: 60 * time.Second,  // 延长心跳间隔
}

// 根据电量动态调整
if battery.Level < 30 {
    agent.DisableSensor("gps")       // 关闭高功耗传感器
    agent.DisableSensor("heart_rate")
}
```

### Q: 如何确保消息可靠?

```go
// IoT 设备使用 QoS 2
config := iot.IoTConfig{
    MQTTQoS: 2,  // 最高可靠性
    ShadowEnabled: true, // 启用设备影子
}

// 关键命令使用确认机制
result := agent.PublishEventWithAck("important_event", data, 30*time.Second)
```

---

## 更多资源

- [用户指南](USER_GUIDE.md) - 完整使用说明
- [API 文档](API.md) - API 参考
- [架构设计](03-ARCHITECTURE_DESIGN.md) - 系统架构
- [GitHub 仓库](https://github.com/ofa/ofa)