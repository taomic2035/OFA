"""
Built-in Skills
Sprint 29: Python Agent SDK
"""

import json
import math
import logging
from typing import Any, Dict, List, Union
from .skills import Skill, SkillInfo

logger = logging.getLogger(__name__)


class EchoSkill(Skill):
    """回显技能"""

    def info(self) -> SkillInfo:
        return SkillInfo(
            id="echo",
            name="Echo",
            description="原样返回输入",
            operations=["echo", "ping"],
            version="1.0.0",
            tags=["utility", "test"],
        )

    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        if operation == "echo":
            return input_data
        elif operation == "ping":
            return {"pong": True, "timestamp": input_data.get("timestamp")}
        return input_data


class TextProcessSkill(Skill):
    """文本处理技能"""

    def info(self) -> SkillInfo:
        return SkillInfo(
            id="text.process",
            name="Text Process",
            description="文本处理操作",
            operations=["uppercase", "lowercase", "reverse", "length",
                       "trim", "split", "replace", "count_words"],
            version="1.0.0",
            tags=["text", "utility"],
        )

    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        text = input_data.get("text", "")

        if operation == "uppercase":
            return text.upper()
        elif operation == "lowercase":
            return text.lower()
        elif operation == "reverse":
            return text[::-1]
        elif operation == "length":
            return len(text)
        elif operation == "trim":
            return text.strip()
        elif operation == "split":
            separator = input_data.get("separator", " ")
            return text.split(separator)
        elif operation == "replace":
            old = input_data.get("old", "")
            new = input_data.get("new", "")
            return text.replace(old, new)
        elif operation == "count_words":
            return len(text.split())

        return text


class CalculatorSkill(Skill):
    """计算器技能"""

    def info(self) -> SkillInfo:
        return SkillInfo(
            id="calculator",
            name="Calculator",
            description="数学计算",
            operations=["add", "sub", "mul", "div", "pow", "sqrt",
                       "mod", "abs", "round", "floor", "ceil", "sin", "cos", "log"],
            version="1.0.0",
            tags=["math", "utility"],
        )

    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        a = input_data.get("a", 0)
        b = input_data.get("b", 0)

        try:
            if operation == "add":
                return a + b
            elif operation == "sub":
                return a - b
            elif operation == "mul":
                return a * b
            elif operation == "div":
                if b == 0:
                    raise ValueError("Division by zero")
                return a / b
            elif operation == "pow":
                return math.pow(a, b)
            elif operation == "sqrt":
                return math.sqrt(a)
            elif operation == "mod":
                return a % b
            elif operation == "abs":
                return abs(a)
            elif operation == "round":
                return round(a, b)
            elif operation == "floor":
                return math.floor(a)
            elif operation == "ceil":
                return math.ceil(a)
            elif operation == "sin":
                return math.sin(a)
            elif operation == "cos":
                return math.cos(a)
            elif operation == "log":
                return math.log(a, b) if b else math.log(a)

        except Exception as e:
            raise ValueError(f"Calculation error: {e}")


class JSONSkill(Skill):
    """JSON处理技能"""

    def info(self) -> SkillInfo:
        return SkillInfo(
            id="json.process",
            name="JSON Process",
            description="JSON数据处理",
            operations=["parse", "stringify", "get_keys", "get_values",
                       "get", "set", "delete", "merge", "validate"],
            version="1.0.0",
            tags=["json", "data"],
        )

    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        data = input_data.get("data")
        path = input_data.get("path", "")

        if operation == "parse":
            if isinstance(data, str):
                return json.loads(data)
            return data

        elif operation == "stringify":
            return json.dumps(data, indent=input_data.get("indent", 2))

        elif operation == "get_keys":
            if isinstance(data, dict):
                return list(data.keys())
            return []

        elif operation == "get_values":
            if isinstance(data, dict):
                return list(data.values())
            return []

        elif operation == "get":
            if path and isinstance(data, dict):
                keys = path.split(".")
                result = data
                for key in keys:
                    if isinstance(result, dict):
                        result = result.get(key)
                    else:
                        return None
                return result
            return data

        elif operation == "set":
            if path and isinstance(data, dict):
                keys = path.split(".")
                value = input_data.get("value")
                target = data
                for key in keys[:-1]:
                    if key not in target:
                        target[key] = {}
                    target = target[key]
                target[keys[-1]] = value
            return data

        elif operation == "delete":
            if path and isinstance(data, dict):
                keys = path.split(".")
                target = data
                for key in keys[:-1]:
                    if isinstance(target, dict):
                        target = target.get(key, {})
                if isinstance(target, dict):
                    target.pop(keys[-1], None)
            return data

        elif operation == "merge":
            other = input_data.get("other", {})
            if isinstance(data, dict) and isinstance(other, dict):
                return {**data, **other}
            return data

        elif operation == "validate":
            try:
                if isinstance(data, str):
                    json.loads(data)
                else:
                    json.dumps(data)
                return {"valid": True}
            except json.JSONDecodeError:
                return {"valid": False, "error": "Invalid JSON"}

        return data


class ListSkill(Skill):
    """列表处理技能"""

    def info(self) -> SkillInfo:
        return SkillInfo(
            id="list.process",
            name="List Process",
            description="列表数据处理",
            operations=["sort", "reverse", "filter", "map", "reduce",
                       "first", "last", "slice", "concat", "unique", "count"],
            version="1.0.0",
            tags=["list", "data"],
        )

    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        data = input_data.get("data", [])

        if operation == "sort":
            reverse = input_data.get("reverse", False)
            return sorted(data, reverse=reverse)

        elif operation == "reverse":
            return list(reversed(data))

        elif operation == "filter":
            condition = input_data.get("condition", "")
            # 简化过滤
            return [item for item in data if item]

        elif operation == "first":
            return data[0] if data else None

        elif operation == "last":
            return data[-1] if data else None

        elif operation == "slice":
            start = input_data.get("start", 0)
            end = input_data.get("end", len(data))
            return data[start:end]

        elif operation == "concat":
            other = input_data.get("other", [])
            return data + other

        elif operation == "unique":
            return list(set(data))

        elif operation == "count":
            return len(data)

        return data


# 技能注册表
BUILTIN_SKILLS = [
    EchoSkill(),
    TextProcessSkill(),
    CalculatorSkill(),
    JSONSkill(),
    ListSkill(),
]


def register_builtin_skills(executor) -> None:
    """注册内置技能"""
    for skill in BUILTIN_SKILLS:
        executor.register(skill)