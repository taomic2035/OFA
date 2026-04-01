"""
OFA Constraint Module
Agent交互约束检查
"""

import json
import logging
import re
from dataclasses import dataclass
from enum import Enum, Flag
from typing import Any, Dict, List, Optional, Set

logger = logging.getLogger(__name__)


class ConstraintType(Flag):
    """约束类型"""
    NONE = 0
    PRIVACY = 1        # 隐私保护
    FINANCIAL = 2      # 财产相关
    SECURITY = 4       # 安全敏感
    AUTH_REQUIRED = 8  # 需要授权
    LOCATION = 16      # 位置信息
    PERSONAL = 32      # 个人信息
    DEVICE = 64        # 设备操作


@dataclass
class ConstraintResult:
    """约束检查结果"""
    allowed: bool
    violated: ConstraintType
    reason: Optional[str] = None
    requires_auth: bool = False
    suggestions: List[str] = None

    def __post_init__(self):
        if self.suggestions is None:
            self.suggestions = []


class ConstraintRule:
    """约束规则"""

    def __init__(
        self,
        name: str,
        constraint_type: ConstraintType,
        action_pattern: str = None,
        data_pattern: str = None,
        check_func: callable = None,
        offline_restricted: bool = False,
        requires_auth: bool = False,
        message: str = ""
    ):
        self.name = name
        self.constraint_type = constraint_type
        self.action_pattern = action_pattern
        self.data_pattern = data_pattern
        self.check_func = check_func
        self.offline_restricted = offline_restricted
        self.requires_auth = requires_auth
        self.message = message


class ConstraintEngine:
    """约束检查引擎"""

    # 默认敏感字段
    SENSITIVE_FIELDS = {
        "idcard", "id_card", "身份证",
        "passport", "护照",
        "phone", "mobile", "电话", "手机",
        "email", "邮箱",
        "address", "地址",
        "bank_account", "银行卡",
        "credit_card", "信用卡",
        "password", "密码",
        "token", "令牌",
        "secret", "密钥",
        "location", "gps", "位置",
    }

    # 默认敏感操作
    SENSITIVE_ACTIONS = {
        "payment", "支付",
        "transfer", "转账",
        "withdraw", "提现",
        "delete_account", "删除账号",
        "change_password", "修改密码",
        "export_data", "导出数据",
        "share", "分享",
        "location_share", "位置分享",
    }

    def __init__(self):
        self._rules: List[ConstraintRule] = []
        self._offline_restricted_actions: Set[str] = set()
        self._custom_sensitive_fields: Set[str] = set()
        self._action_hooks: Dict[str, List[callable]] = {}
        self._load_default_rules()

    def _load_default_rules(self):
        """加载默认规则"""
        # 财务操作规则
        self.add_rule(ConstraintRule(
            name="financial_operations",
            constraint_type=ConstraintType.FINANCIAL,
            action_pattern=r"(payment|transfer|withdraw|pay)",
            offline_restricted=True,
            requires_auth=True,
            message="Financial operations require online mode and authorization"
        ))

        # 隐私数据规则
        self.add_rule(ConstraintRule(
            name="privacy_data",
            constraint_type=ConstraintType.PRIVACY,
            data_pattern=r"(idcard|id_card|身份证|passport|护照)",
            offline_restricted=False,
            requires_auth=False,
            message="Data contains sensitive personal information"
        ))

        # 位置信息规则
        self.add_rule(ConstraintRule(
            name="location_data",
            constraint_type=ConstraintType.LOCATION,
            data_pattern=r"(location|gps|latitude|longitude|经纬度)",
            offline_restricted=False,
            requires_auth=True,
            message="Location data sharing requires authorization"
        ))

        # 安全操作规则
        self.add_rule(ConstraintRule(
            name="security_operations",
            constraint_type=ConstraintType.SECURITY,
            action_pattern=r"(delete|password|auth|login|logout)",
            offline_restricted=True,
            requires_auth=True,
            message="Security operations require online mode and authorization"
        ))

        # 设备操作规则
        self.add_rule(ConstraintRule(
            name="device_operations",
            constraint_type=ConstraintType.DEVICE,
            action_pattern=r"(shutdown|reboot|factory_reset|install|uninstall)",
            offline_restricted=False,
            requires_auth=True,
            message="Device operations require authorization"
        ))

    def add_rule(self, rule: ConstraintRule):
        """添加规则"""
        self._rules.append(rule)

        if rule.offline_restricted and rule.action_pattern:
            pattern = rule.action_pattern.replace(r"(", "").replace(r")", "")
            for action in pattern.split("|"):
                self._offline_restricted_actions.add(action.strip())

    def remove_rule(self, name: str):
        """移除规则"""
        self._rules = [r for r in self._rules if r.name != name]

    def check(
        self,
        action: str,
        data: Any = None,
        offline_mode: bool = False,
        user_context: Dict[str, Any] = None
    ) -> ConstraintResult:
        """检查约束"""
        result = ConstraintResult(
            allowed=True,
            violated=ConstraintType.NONE,
        )

        user_context = user_context or {}

        # 1. 检查离线受限操作
        if offline_mode:
            for restricted in self._offline_restricted_actions:
                if action.lower().find(restricted) >= 0:
                    result.allowed = False
                    result.violated |= ConstraintType.FINANCIAL | ConstraintType.SECURITY
                    result.reason = f"Action '{action}' requires online mode"
                    result.suggestions.append("Connect to network or use alternative offline action")
                    return result

        # 2. 应用自定义规则
        for rule in self._rules:
            rule_result = self._apply_rule(rule, action, data, offline_mode, user_context)
            if not rule_result.allowed:
                return rule_result

        # 3. 检查敏感数据
        if data:
            data_result = self._check_sensitive_data(data)
            if data_result.violated:
                result.allowed = False
                result.violated |= data_result.violated
                result.reason = data_result.reason
                result.requires_auth = bool(result.violated & ConstraintType.AUTH_REQUIRED)
                return result

        # 4. 执行钩子
        hooks = self._action_hooks.get(action, [])
        for hook in hooks:
            try:
                hook_result = hook(action, data, user_context)
                if hook_result and not hook_result.allowed:
                    return hook_result
            except Exception as e:
                logger.error(f"Constraint hook error: {e}")

        return result

    def _apply_rule(
        self,
        rule: ConstraintRule,
        action: str,
        data: Any,
        offline_mode: bool,
        user_context: Dict[str, Any]
    ) -> ConstraintResult:
        """应用单个规则"""
        result = ConstraintResult(allowed=True, violated=ConstraintType.NONE)

        # 检查操作模式
        if rule.action_pattern:
            if not re.search(rule.action_pattern, action, re.IGNORECASE):
                return result

        # 检查数据模式
        if rule.data_pattern and data:
            data_str = json.dumps(data) if isinstance(data, (dict, list)) else str(data)
            if not re.search(rule.data_pattern, data_str, re.IGNORECASE):
                return result

        # 自定义检查函数
        if rule.check_func:
            try:
                custom_result = rule.check_func(action, data, user_context)
                if custom_result is False:
                    result.allowed = False
            except Exception as e:
                logger.error(f"Custom check error: {e}")

        # 离线限制
        if rule.offline_restricted and offline_mode:
            result.allowed = False
            result.violated |= rule.constraint_type
            result.reason = rule.message
            result.suggestions.append("Switch to online mode")
            return result

        # 授权要求
        if rule.requires_auth:
            if not user_context.get("authorized", False):
                result.allowed = False
                result.violated |= ConstraintType.AUTH_REQUIRED
                result.reason = rule.message
                result.requires_auth = True
                result.suggestions.append("Request authorization from user")

        return result

    def _check_sensitive_data(self, data: Any) -> ConstraintResult:
        """检查敏感数据"""
        result = ConstraintResult(allowed=True, violated=ConstraintType.NONE)

        data_str = json.dumps(data) if isinstance(data, (dict, list)) else str(data)
        data_lower = data_str.lower()

        # 检查所有敏感字段
        all_sensitive = self.SENSITIVE_FIELDS | self._custom_sensitive_fields

        for field in all_sensitive:
            if field in data_lower:
                if field in {"idcard", "id_card", "身份证", "passport", "护照"}:
                    result.violated |= ConstraintType.PRIVACY
                    result.reason = "Data contains identity information"
                elif field in {"bank_account", "credit_card", "银行卡", "信用卡"}:
                    result.violated |= ConstraintType.FINANCIAL
                    result.reason = "Data contains financial information"
                elif field in {"location", "gps", "位置"}:
                    result.violated |= ConstraintType.LOCATION
                    result.reason = "Data contains location information"
                elif field in {"password", "token", "secret", "密码", "密钥"}:
                    result.violated |= ConstraintType.SECURITY
                    result.reason = "Data contains security credentials"

        return result

    def add_sensitive_field(self, field: str):
        """添加自定义敏感字段"""
        self._custom_sensitive_fields.add(field.lower())

    def remove_sensitive_field(self, field: str):
        """移除敏感字段"""
        self._custom_sensitive_fields.discard(field.lower())

    def register_action_hook(self, action: str, hook: callable):
        """注册操作钩子"""
        if action not in self._action_hooks:
            self._action_hooks[action] = []
        self._action_hooks[action].append(hook)

    def unregister_action_hook(self, action: str, hook: callable):
        """移除操作钩子"""
        if action in self._action_hooks:
            self._action_hooks[action] = [h for h in self._action_hooks[action] if h != hook]

    def get_rules_summary(self) -> List[Dict[str, Any]]:
        """获取规则摘要"""
        return [
            {
                "name": r.name,
                "type": r.constraint_type.name,
                "offline_restricted": r.offline_restricted,
                "requires_auth": r.requires_auth,
            }
            for r in self._rules
        ]


class ConstraintClient:
    """约束检查客户端"""

    def __init__(self, engine: ConstraintEngine = None):
        self.engine = engine or ConstraintEngine()
        self._offline_mode = False
        self._user_context: Dict[str, Any] = {}

    def set_offline_mode(self, offline: bool):
        """设置离线模式"""
        self._offline_mode = offline

    def set_user_context(self, context: Dict[str, Any]):
        """设置用户上下文"""
        self._user_context = context

    def check(self, action: str, data: Any = None) -> ConstraintResult:
        """执行约束检查"""
        return self.engine.check(
            action,
            data,
            offline_mode=self._offline_mode,
            user_context=self._user_context
        )

    def check_and_raise(self, action: str, data: Any = None) -> bool:
        """检查并抛出异常"""
        result = self.check(action, data)
        if not result.allowed:
            raise ConstraintViolationError(
                result.violated,
                result.reason or "Operation not allowed"
            )
        return True

    def is_allowed(self, action: str, data: Any = None) -> bool:
        """快速检查是否允许"""
        return self.check(action, data).allowed

    def add_rule(self, rule: ConstraintRule):
        """添加规则"""
        self.engine.add_rule(rule)

    def get_offline_restricted_actions(self) -> Set[str]:
        """获取离线受限操作"""
        return self.engine._offline_restricted_actions.copy()


class ConstraintViolationError(Exception):
    """约束违反异常"""

    def __init__(self, violated_type: ConstraintType, message: str):
        self.violated_type = violated_type
        super().__init__(message)