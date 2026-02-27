# 行为 / 验收规格 (Behavior Specs)

本目录存放**可验收的行为场景**，用于 BDD/验收测试。

## 格式选项

1. **Gherkin**（`.feature`）：可与 Cucumber、Behave 等工具对接。
2. **Markdown**：Given/When/Then 或类似结构，便于阅读与手工验收。

## 示例（Gherkin）

```gherkin
Feature: 比赛录像列表
  Scenario: 用户按最近 N 天查询录像
    When 用户请求 GET /replays?days=7
    Then 返回该时间范围内的录像列表
    And 每条记录仅包含 title、url、date
    And 响应中不得包含比赛比分
```

实现或自动化测试应覆盖这些场景，保证行为与 spec 一致。**领域约束**：不透露比赛比分（见 `specs/domain/replay.md`）。
