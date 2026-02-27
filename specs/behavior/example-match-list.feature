# 行为规格：与 specs/api/openapi.yaml 中的 /replays 契约一致，可作为验收场景参考
# 领域约束：不透露比赛比分（见 specs/domain/replay.md）
Feature: 巴塞罗那比赛录像列表
  As a 用户
  I want to 按最近 N 天查询巴塞罗那比赛录像
  So that 我能快速看到近期录像链接（且不被剧透比分）

  Scenario: 响应不得包含比赛比分
    When 用户请求 GET /replays?days=7
    Then 返回状态 200
    And 每条记录仅包含 "title"、"url"、"date"
    And 响应中不得出现比分、结果、进球数等剧透信息

  Scenario: 健康检查
    When 用户请求 GET /health
    Then 返回状态 200
    And 响应 JSON 包含 "status" 且值为 "ok"

  Scenario: 查询最近 7 天录像（默认）
    When 用户请求 GET /replays
    Then 返回状态 200
    And 响应 JSON 包含 "items" 数组
    And 每条记录包含 "title"、"url"、"date"

  Scenario: 查询最近 14 天录像
    When 用户请求 GET /replays?days=14
    Then 返回状态 200
    And 响应 JSON 包含 "items" 数组
    And 每条记录的 "date" 在最近 14 天内

  Scenario: 参数 days 超出范围时返回 400
    When 用户请求 GET /replays?days=0
    Then 返回状态 400
    When 用户请求 GET /replays?days=31
    Then 返回状态 400
