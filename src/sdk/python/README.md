# OFA Python SDK

提供Python语言的OFA Agent SDK，用于快速开发和集成Agent。

## 安装

```bash
pip install ofa-sdk
```

## 快速开始

```python
import asyncio
from ofa import Agent, Skill

# 定义技能
class HelloSkill(Skill):
    @property
    def id(self):
        return "hello"

    @property
    def name(self):
        return "Hello World"

    async def execute(self, input_data: dict) -> dict:
        name = input_data.get("name", "World")
        return {"message": f"Hello, {name}!"}

# 创建Agent
async def main():
    agent = Agent(
        center_addr="localhost:9090",
        name="python-agent",
        agent_type="full"
    )

    # 注册技能
    agent.register_skill(HelloSkill())

    # 连接到Center
    await agent.connect()

    print(f"Agent started: {agent.id}")

    # 保持运行
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        await agent.disconnect()

if __name__ == "__main__":
    asyncio.run(main())
```

## API参考

### Agent

Agent客户端类。

#### 构造函数

```python
Agent(center_addr: str, name: str, agent_type: str = "full")
```

#### 方法

- `async connect()` - 连接到Center
- `async disconnect()` - 断开连接
- `register_skill(skill: Skill)` - 注册技能
- `async send_message(to_agent: str, action: str, payload: dict)` - 发送消息

### Skill

技能基类。

#### 属性

- `id: str` - 技能ID
- `name: str` - 技能名称
- `version: str` - 版本号
- `category: str` - 分类

#### 方法

- `async execute(input_data: dict) -> dict` - 执行技能

## 示例

### 计算器技能

```python
from ofa import Skill

class CalculatorSkill(Skill):
    @property
    def id(self):
        return "calculator"

    @property
    def name(self):
        return "Calculator"

    async def execute(self, input_data: dict) -> dict:
        operation = input_data.get("operation")
        a = input_data.get("a", 0)
        b = input_data.get("b", 0)

        if operation == "add":
            result = a + b
        elif operation == "sub":
            result = a - b
        elif operation == "mul":
            result = a * b
        elif operation == "div":
            result = a / b
        else:
            raise ValueError(f"Unknown operation: {operation}")

        return {"result": result}
```

### HTTP请求技能

```python
import aiohttp
from ofa import Skill

class HttpRequestSkill(Skill):
    @property
    def id(self):
        return "http.request"

    @property
    def name(self):
        return "HTTP Request"

    async def execute(self, input_data: dict) -> dict:
        url = input_data.get("url")
        method = input_data.get("method", "GET")
        headers = input_data.get("headers", {})
        data = input_data.get("data")

        async with aiohttp.ClientSession() as session:
            async with session.request(
                method, url, headers=headers, json=data
            ) as response:
                return {
                    "status": response.status,
                    "body": await response.text()
                }
```