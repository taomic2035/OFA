# OFA Skills Directory

This directory contains reusable skills that can be deployed to agents.

## Directory Structure

```
skills/
├── core/           # Core skills (built-in)
│   ├── text/       # Text processing
│   ├── json/       # JSON processing
│   └── calc/       # Calculator
├── ai/             # AI-related skills
│   ├── llm/        # LLM integration
│   ├── vision/     # Image/video processing
│   └── audio/      # Audio processing
├── data/           # Data processing skills
│   ├── transform/  # Data transformation
│   ├── validate/   # Data validation
│   └── analyze/    # Data analysis
├── io/             # I/O skills
│   ├── file/       # File operations
│   ├── network/    # Network operations
│   └── database/   # Database operations
└── custom/         # Custom skills
```

## Skill Format

Each skill should include:
- `skill.yaml` - Skill metadata
- `main.py` or `main.go` - Implementation
- `README.md` - Documentation
- `test/` - Tests

## Example skill.yaml

```yaml
id: text.process
name: Text Process
version: 1.0.0
category: text
description: Process text with various operations
author: OFA Team
license: MIT

operations:
  - uppercase
  - lowercase
  - reverse
  - length

input_schema:
  type: object
  properties:
    text:
      type: string
      description: Input text
    operation:
      type: string
      enum: [uppercase, lowercase, reverse, length]
      description: Operation to perform

output_schema:
  type: object
  properties:
    result:
      type: string
      description: Processed text

requirements:
  - python >= 3.8

tags:
  - text
  - processing
  - utility
```