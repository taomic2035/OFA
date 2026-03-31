"""
OFA Python Agent Example
Sprint 29: Python Agent SDK
"""

import asyncio
import logging
from ofa_agent import (
    OFAAgent, AgentConfig,
    SkillExecutor, FunctionSkill,
    register_builtin_skills,
)

logging.basicConfig(level=logging.INFO)


async def main():
    # 创建配置
    config = AgentConfig(
        agent_id="python-demo-001",
        name="Python Demo Agent",
        center_url="localhost:9090",
        connection_type="http",
        skills=["echo", "text.process", "calculator", "json.process"],
    )

    # 创建Agent
    agent = OFAAgent(config)

    # 创建技能执行器
    executor = SkillExecutor()

    # 注册内置技能
    register_builtin_skills(executor)

    # 注册自定义技能
    async def custom_handler(operation, input_data):
        return {"processed": True, "input": input_data}

    custom_skill = FunctionSkill(
        skill_id="custom.process",
        name="Custom Processor",
        operations={"process": custom_handler},
        description="自定义处理器",
    )
    executor.register(custom_skill)

    # 注册技能到Agent
    for skill in executor.list_skills():
        agent.register_skill(skill.id, executor.execute)

    # 连接Center
    # await agent.connect()

    print(f"Agent ID: {agent.info.id}")
    print(f"Agent Name: {agent.info.name}")
    print(f"Skills: {agent.info.skills}")

    # 测试技能执行
    result = await executor.execute("echo", "echo", {"message": "Hello OFA!"})
    print(f"Echo result: {result}")

    result = await executor.execute("text.process", "uppercase", {"text": "hello world"})
    print(f"Text result: {result}")

    result = await executor.execute("calculator", "add", {"a": 10, "b": 20})
    print(f"Calculator result: {result}")

    result = await executor.execute("json.process", "get_keys", {"data": {"a": 1, "b": 2, "c": 3}})
    print(f"JSON result: {result}")

    # 断开连接
    # await agent.disconnect()


if __name__ == "__main__":
    asyncio.run(main())