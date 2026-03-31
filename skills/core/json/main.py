"""
JSON Process Skill Implementation
"""

import json
from typing import Dict, Any, List


def execute(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """Execute JSON processing operation"""

    json_str = input_data.get("json_str", "")
    operation = input_data.get("operation", "pretty")
    options = input_data.get("options", {})

    if not json_str and operation != "merge":
        return {"error": "No JSON provided"}

    try:
        data = json.loads(json_str) if json_str else {}
    except json.JSONDecodeError as e:
        if operation == "validate":
            return {"result": False, "error": str(e)}
        return {"error": f"Invalid JSON: {e}"}

    def get_keys(obj: Any, prefix: str = "") -> List[str]:
        keys = []
        if isinstance(obj, dict):
            for k, v in obj.items():
                full_key = f"{prefix}.{k}" if prefix else k
                keys.append(full_key)
                keys.extend(get_keys(v, full_key))
        elif isinstance(obj, list):
            for i, v in enumerate(obj):
                full_key = f"{prefix}[{i}]"
                keys.extend(get_keys(v, full_key))
        return keys

    def get_values(obj: Any) -> List[Any]:
        values = []
        if isinstance(obj, dict):
            for v in obj.values():
                if isinstance(v, (dict, list)):
                    values.extend(get_values(v))
                else:
                    values.append(v)
        elif isinstance(obj, list):
            for v in obj:
                if isinstance(v, (dict, list)):
                    values.extend(get_values(v))
                else:
                    values.append(v)
        return values

    def extract_path(obj: Any, path: str) -> Any:
        parts = path.replace("[", ".").replace("]", "").split(".")
        for part in parts:
            if not part:
                continue
            if isinstance(obj, dict):
                obj = obj.get(part)
            elif isinstance(obj, list):
                try:
                    obj = obj[int(part)]
                except (ValueError, IndexError):
                    return None
            else:
                return None
        return obj

    operations = {
        "get_keys": lambda d: get_keys(d),
        "get_values": lambda d: get_values(d),
        "pretty": lambda d: json.dumps(d, indent=options.get("indent", 2)),
        "validate": lambda d: True,
        "merge": lambda d: json.loads(options.get("other_json", "{}")) | d if isinstance(d, dict) else d,
        "extract": lambda d: extract_path(d, options.get("path", "")),
    }

    if operation not in operations:
        return {"error": f"Unknown operation: {operation}"}

    try:
        result = operations[operation](data)
        return {"result": result, "operation": operation}
    except Exception as e:
        return {"error": str(e)}


# For testing
if __name__ == "__main__":
    test_json = '{"name": "test", "value": 123, "nested": {"a": 1, "b": 2}}'

    tests = [
        {"json_str": test_json, "operation": "get_keys"},
        {"json_str": test_json, "operation": "get_values"},
        {"json_str": test_json, "operation": "pretty"},
        {"json_str": test_json, "operation": "validate"},
        {"json_str": test_json, "operation": "extract", "options": {"path": "nested.a"}},
        {"json_str": '{"a": 1}', "operation": "merge", "options": {"other_json": '{"b": 2}'}},
    ]

    for test in tests:
        result = execute(test)
        print(f"Input: {test}")
        print(f"Output: {result}")
        print()