"""
Skills Module
Sprint 29: Python Agent SDK
"""

import asyncio
import logging
from abc import ABC, abstractmethod
from typing import Any, Dict, List, Optional, Callable
from dataclasses import dataclass, field

logger = logging.getLogger(__name__)


@dataclass
class SkillInfo:
    """技能信息"""
    id: str
    name: str
    description: str = ""
    operations: List[str] = field(default_factory=list)
    version: str = "1.0.0"
    author: str = "OFA"
    tags: List[str] = field(default_factory=list)


class Skill(ABC):
    """技能基类"""

    @abstractmethod
    def info(self) -> SkillInfo:
        """返回技能信息"""
        pass

    @abstractmethod
    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        """执行技能"""
        pass

    def validate_operation(self, operation: str) -> bool:
        """验证操作是否有效"""
        return operation in self.info().operations


class SkillExecutor:
    """技能执行器"""

    def __init__(self):
        self._skills: Dict[str, Skill] = {}
        self._stats: Dict[str, Dict] = {}

    def register(self, skill: Skill) -> None:
        """注册技能"""
        info = skill.info()
        self._skills[info.id] = skill
        self._stats[info.id] = {
            "invocations": 0,
            "successes": 0,
            "failures": 0,
            "total_time": 0.0,
        }
        logger.info(f"Skill registered: {info.id}")

    def unregister(self, skill_id: str) -> None:
        """注销技能"""
        self._skills.pop(skill_id, None)
        self._stats.pop(skill_id, None)

    def get_skill(self, skill_id: str) -> Optional[Skill]:
        """获取技能"""
        return self._skills.get(skill_id)

    def list_skills(self) -> List[SkillInfo]:
        """列出所有技能"""
        return [skill.info() for skill in self._skills.values()]

    async def execute(self, skill_id: str, operation: str,
                      input_data: Dict[str, Any]) -> Dict[str, Any]:
        """执行技能"""
        skill = self._skills.get(skill_id)
        if not skill:
            return {
                "success": False,
                "error": f"Skill not found: {skill_id}",
            }

        if not skill.validate_operation(operation):
            return {
                "success": False,
                "error": f"Invalid operation: {operation}",
            }

        stats = self._stats[skill_id]
        stats["invocations"] += 1

        import time
        start = time.time()

        try:
            result = await skill.execute(operation, input_data)
            stats["successes"] += 1
            stats["total_time"] += time.time() - start

            return {
                "success": True,
                "result": result,
                "skill_id": skill_id,
                "operation": operation,
            }

        except Exception as e:
            stats["failures"] += 1
            logger.error(f"Skill execution failed: {e}")

            return {
                "success": False,
                "error": str(e),
                "skill_id": skill_id,
                "operation": operation,
            }

    def get_stats(self, skill_id: str) -> Optional[Dict]:
        """获取统计"""
        return self._stats.get(skill_id)

    def get_all_stats(self) -> Dict[str, Dict]:
        """获取所有统计"""
        return self._stats.copy()


class SkillRegistry:
    """技能注册表"""

    def __init__(self):
        self._registry: Dict[str, SkillInfo] = {}
        self._categories: Dict[str, List[str]] = {}

    def register(self, skill: Skill) -> None:
        """注册技能"""
        info = skill.info()
        self._registry[info.id] = info

        # 添加到分类
        for tag in info.tags:
            if tag not in self._categories:
                self._categories[tag] = []
            if info.id not in self._categories[tag]:
                self._categories[tag].append(info.id)

    def get(self, skill_id: str) -> Optional[SkillInfo]:
        """获取技能信息"""
        return self._registry.get(skill_id)

    def search(self, query: str) -> List[SkillInfo]:
        """搜索技能"""
        query = query.lower()
        results = []

        for info in self._registry.values():
            if query in info.name.lower() or \
               query in info.description.lower() or \
               query in info.id.lower():
                results.append(info)

        return results

    def by_category(self, category: str) -> List[SkillInfo]:
        """按分类获取"""
        skill_ids = self._categories.get(category, [])
        return [self._registry[id] for id in skill_ids if id in self._registry]

    def list_all(self) -> List[SkillInfo]:
        """列出所有"""
        return list(self._registry.values())


class FunctionSkill(Skill):
    """函数技能包装器"""

    def __init__(self, skill_id: str, name: str,
                 operations: Dict[str, Callable],
                 description: str = ""):
        self._id = skill_id
        self._name = name
        self._description = description
        self._operations = operations

    def info(self) -> SkillInfo:
        return SkillInfo(
            id=self._id,
            name=self._name,
            description=self._description,
            operations=list(self._operations.keys()),
        )

    async def execute(self, operation: str, input_data: Dict[str, Any]) -> Any:
        handler = self._operations.get(operation)
        if not handler:
            raise ValueError(f"Operation not found: {operation}")

        # 支持同步和异步
        if asyncio.iscoroutinefunction(handler):
            return await handler(input_data)
        else:
            return handler(input_data)