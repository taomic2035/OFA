"""
OFA Python SDK 离线模式示例
演示 L1-L4 离线能力
"""

import asyncio
import time
from ofa_agent import (
    OFAAgent,
    AgentConfig,
    OfflineManager,
    OfflineLevel,
    P2PClient,
    P2PDiscovery,
    ConstraintClient,
    ConstraintType,
)


def demo_offline_scheduler():
    """演示本地调度器"""
    print("\n=== 本地调度器示例 ===")

    # 创建离线管理器 (L1 完全离线)
    manager = OfflineManager(offline_level=OfflineLevel.L1)
    manager.start()

    # 注册离线技能
    def text_uppercase(input_data):
        return input_data.upper() if isinstance(input_data, str) else input_data

    def calculate(data):
        return eval(str(data))  # 注意: 生产环境应使用更安全的方式

    manager.register_skill("text.uppercase", text_uppercase)
    manager.register_skill("calculate", calculate)

    # 同步执行
    result = manager.execute_sync("text.uppercase", "hello world")
    print(f"结果: {result}")

    # 缓存数据
    manager.cache_data("user:123", {"name": "张三", "age": 30})
    cached = manager.get_cached("user:123")
    print(f"缓存: {cached}")

    # 统计
    stats = manager.get_stats()
    print(f"统计: {stats}")

    manager.stop()


def demo_p2p():
    """演示 P2P 通信"""
    print("\n=== P2P 通信示例 ===")

    # 创建两个 P2P 客户端
    client1 = P2PClient(agent_id="agent-1")
    client2 = P2PClient(agent_id="agent-2")

    client1.start()
    client2.start()

    # 添加消息处理器
    def on_message(msg):
        print(f"收到消息: {msg.from_id} -> {msg.data}")

    client2.on_message(on_message)

    # 手动添加设备
    from ofa_agent import PeerInfo

    peer2 = PeerInfo(
        id="agent-2",
        name="Agent 2",
        address="127.0.0.1",
        port=client2.port,
    )
    client1.add_peer(peer2)

    # 发送消息
    client1.send("agent-2", {"text": "Hello from agent-1!"})

    # 广播
    results = client1.broadcast({"event": "test", "data": "broadcast message"})
    print(f"广播结果: {results}")

    # 统计
    print(f"Agent-1 统计: {client1.get_stats()}")

    client1.stop()
    client2.stop()


def demo_constraint():
    """演示约束检查"""
    print("\n=== 约束检查示例 ===")

    client = ConstraintClient()

    # 检查普通操作
    result = client.check("text.process", {"text": "hello"})
    print(f"普通操作: allowed={result.allowed}")

    # 检查敏感操作 (支付)
    result = client.check("payment.send", {"amount": 100})
    print(f"支付操作: allowed={result.allowed}, reason={result.reason}")

    # 检查敏感数据
    result = client.check("save_data", {"idcard": "123456789"})
    print(f"敏感数据: allowed={result.allowed}, violated={result.violated}")

    # 离线模式下的限制
    client.set_offline_mode(True)
    result = client.check("payment.send", {"amount": 100})
    print(f"离线支付: allowed={result.allowed}, reason={result.reason}")

    # 添加自定义规则
    from ofa_agent import ConstraintRule

    client.add_rule(ConstraintRule(
        name="custom_night_restriction",
        constraint_type=ConstraintType.SECURITY,
        action_pattern=r"delete",
        check_func=lambda action, data, ctx: False if ctx.get("night_mode") else True,
        message="Delete operations not allowed at night"
    ))

    client.set_user_context({"night_mode": True})
    result = client.check("delete_file", {"file": "test.txt"})
    print(f"夜间删除: allowed={result.allowed}, reason={result.reason}")


def demo_full_offline_agent():
    """演示完整的离线 Agent"""
    print("\n=== 完整离线 Agent 示例 ===")

    # 创建配置
    config = AgentConfig(
        name="Offline Agent Demo",
        center_url="localhost:9090",
    )

    # 创建 Agent
    agent = OFAAgent(config)

    # 创建离线管理器
    offline = OfflineManager(offline_level=OfflineLevel.L1)
    offline.start()

    # 注册离线技能
    offline.register_skill("echo", lambda x: x)
    offline.register_skill("time", lambda x: time.strftime("%Y-%m-%d %H:%M:%S"))
    offline.register_skill("math.add", lambda x: x.get("a", 0) + x.get("b", 0))

    # 执行本地任务
    task_id = offline.execute_local("echo", "Hello Offline!")
    print(f"任务提交: {task_id}")

    # 等待完成
    time.sleep(0.5)
    task = offline.scheduler.get_task(task_id)
    if task:
        print(f"任务状态: {task.status.value}, 结果: {task.output_data}")

    # 批量任务
    for i in range(5):
        offline.execute_local("math.add", {"a": i, "b": i * 2})

    time.sleep(1)
    print(f"统计: {offline.get_stats()}")

    offline.stop()
    print("离线 Agent 示例完成")


def main():
    print("OFA Python SDK 离线模式演示")
    print("=" * 50)

    # 演示各个模块
    demo_offline_scheduler()
    demo_constraint()
    demo_p2p()
    demo_full_offline_agent()

    print("\n所有演示完成!")


if __name__ == "__main__":
    main()