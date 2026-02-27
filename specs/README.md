# Specifications (Spec-Driven)

本目录是所有**规格说明**的单一事实来源（Single Source of Truth）。先写 spec，再实现代码。

## 目录结构

```
specs/
├── README.md           # 本文件：规范约定与工作流
├── api/                # API 规格（OpenAPI 3.x）
│   └── openapi.yaml    # 或 .json
├── domain/             # 领域模型与业务规则（可选）
│   └── *.md
└── behavior/           # 行为/验收规格（可选，BDD 风格）
    └── *.feature 或 *.md
```

## 工作流

1. **先写 spec**：在 `specs/` 下新增或修改规格（API 路径、请求/响应、业务规则、场景）。
2. **评审**：确认 spec 无误后再动手写实现。
3. **实现**：代码以满足 spec 为目标；测试可基于 spec 生成或手写。
4. **回归**：每次改动后用 spec 做契约测试或验收测试，确保实现与 spec 一致。

## 规范格式

| 类型     | 格式/工具        | 用途           |
|----------|------------------|----------------|
| HTTP API | OpenAPI 3.x      | 接口契约、生成文档/客户端 |
| 领域规则 | Markdown / 结构化文档 | 业务规则、领域模型 |
| 行为场景 | Gherkin (.feature) 或 Markdown | 验收条件、BDD |

## 与实现的关系

- **API**：可根据 `specs/api/openapi.yaml` 做契约测试、生成类型或 mock。
- **行为**：可将 `specs/behavior/` 中的场景转为自动化验收测试。

新增功能时，请先在对应子目录下补充或更新 spec，再在代码库中实现。
