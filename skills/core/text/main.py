"""
Text Process Skill Implementation
"""

import json
from typing import Dict, Any


def execute(input_data: Dict[str, Any]) -> Dict[str, Any]:
    """Execute text processing operation"""

    text = input_data.get("text", "")
    operation = input_data.get("operation", "uppercase")
    options = input_data.get("options", {})

    if not text:
        return {"error": "No text provided"}

    operations = {
        "uppercase": lambda t: t.upper(),
        "lowercase": lambda t: t.lower(),
        "reverse": lambda t: t[::-1],
        "length": lambda t: len(t),
        "trim": lambda t: t.strip(),
        "split": lambda t: t.split(options.get("separator", " ")),
        "replace": lambda t: t.replace(
            options.get("old", ""),
            options.get("new", "")
        ),
    }

    if operation not in operations:
        return {"error": f"Unknown operation: {operation}"}

    try:
        result = operations[operation](text)
        return {"result": result, "operation": operation}
    except Exception as e:
        return {"error": str(e)}


# For testing
if __name__ == "__main__":
    # Test cases
    tests = [
        {"text": "hello", "operation": "uppercase"},
        {"text": "HELLO", "operation": "lowercase"},
        {"text": "hello", "operation": "reverse"},
        {"text": "hello world", "operation": "length"},
        {"text": "  hello  ", "operation": "trim"},
        {"text": "hello world", "operation": "split"},
        {"text": "hello world", "operation": "replace", "options": {"old": "world", "new": "there"}},
    ]

    for test in tests:
        result = execute(test)
        print(f"Input: {test}")
        print(f"Output: {result}")
        print()