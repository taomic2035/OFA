#!/usr/bin/env python3
"""
OFA Agent Example
"""

import asyncio
import sys
import os

# Add parent directory to path
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from ofa import Agent, Skill


class TextProcessSkill(Skill):
    """Text processing skill"""

    @property
    def id(self):
        return "text.process"

    @property
    def name(self):
        return "Text Process"

    @property
    def category(self):
        return "text"

    async def execute(self, input_data):
        text = input_data.get("text", "")
        operation = input_data.get("operation", "uppercase")

        if operation == "uppercase":
            result = text.upper()
        elif operation == "lowercase":
            result = text.lower()
        elif operation == "reverse":
            result = text[::-1]
        elif operation == "length":
            result = str(len(text))
        else:
            raise ValueError(f"Unknown operation: {operation}")

        return {"result": result}


class CalculatorSkill(Skill):
    """Calculator skill"""

    @property
    def id(self):
        return "calculator"

    @property
    def name(self):
        return "Calculator"

    @property
    def category(self):
        return "math"

    async def execute(self, input_data):
        operation = input_data.get("operation", "add")
        a = input_data.get("a", 0)
        b = input_data.get("b", 0)

        operations = {
            "add": lambda x, y: x + y,
            "sub": lambda x, y: x - y,
            "mul": lambda x, y: x * y,
            "div": lambda x, y: x / y if y != 0 else float("inf"),
        }

        if operation not in operations:
            raise ValueError(f"Unknown operation: {operation}")

        result = operations[operation](a, b)
        return {"result": result}


async def main():
    # Get configuration from environment
    center_addr = os.environ.get("CENTER_ADDR", "localhost:9090")
    agent_name = os.environ.get("AGENT_NAME", "python-agent")

    # Create agent
    agent = Agent(
        center_addr=center_addr,
        name=agent_name,
        agent_type="full"
    )

    # Register skills
    agent.register_skill(TextProcessSkill())
    agent.register_skill(CalculatorSkill())

    # Set message handler
    async def handle_message(msg):
        print(f"Received message: {msg.action} from {msg.from_agent}")
        if msg.action == "ping":
            await agent.send_message(msg.from_agent, "pong", {"reply_to": msg.msg_id})

    agent.set_message_handler(handle_message)

    # Connect to Center
    try:
        await agent.connect()
        print(f"Agent started: {agent.id}")
        print(f"Registered skills: text.process, calculator")

        # Keep running
        while True:
            await asyncio.sleep(1)

    except KeyboardInterrupt:
        print("\nShutting down...")
    except Exception as e:
        print(f"Error: {e}")
    finally:
        await agent.disconnect()


if __name__ == "__main__":
    asyncio.run(main())