# SoccerTools

以 **Spec-Driven（规格先行）** 方式构建的足球相关工具：从 [直播吧](https://www.zhibo8.com/) 爬取巴塞罗那比赛录像信息，提供 HTTP API 查询最近 N 天的录像列表。

## 技术栈

- **Go 1.21+**
- 规格：OpenAPI 3（`specs/api/`）、行为场景（`specs/behavior/`）

## 项目结构

```
SoccerTools/
├── cmd/server/          # 服务入口，定时爬取 + HTTP API + Web 页
│   └── static/          # 前端页面（embed 打包）
├── internal/
│   ├── crawler/         # 直播吧录像页解析（巴萨相关）
│   ├── model/           # 数据模型
│   └── store/           # 内存存储，按最近 N 天过滤
├── specs/               # 规格（先行）
│   ├── api/             # OpenAPI 契约
│   ├── behavior/        # 行为/验收场景
│   └── domain/           # 领域说明
├── go.mod
└── README.md
```

## 快速开始

### 运行服务

```bash
go run ./cmd/server
```

默认监听 `http://localhost:3000`，启动后会立即爬取一次，之后每 5 秒定时爬取。

### Web 页面

浏览器打开 **http://localhost:3000** 可进入「巴萨比赛录像」查询页：

- **查询**：选择最近 7 / 14 / 30 天，点击「查询」或切换天数自动加载。
- **刷新数据**：点击「刷新数据」会立即从直播吧爬取一次并更新列表。

### API

| 接口 | 方法 | 说明 |
|------|------|------|
| `/` | GET | 巴萨录像查询页（HTML） |
| `/health` | GET | 健康检查，返回 `{"status":"ok"}` |
| `/replays` | GET | 巴塞罗那比赛录像列表 |
| `/replays/refresh` | POST | 立即爬取一次，更新数据 |

**查询最近 N 天录像：**

```bash
# 默认最近 7 天
curl http://localhost:3000/replays

# 指定天数（1–30）
curl "http://localhost:3000/replays?days=14"
```

**响应示例：**

```json
{
  "items": [
    {
      "title": "巴塞罗那vs莱万特",
      "url": "https://www.zhibo8.com/zuqiu/2026/0223-match1730577v-luxiang.htm",
      "date": "2026-02-23"
    }
  ]
}
```

API 契约见 [specs/api/openapi.yaml](specs/api/openapi.yaml)。

## Spec-Driven 约定

1. **先写规格，再写实现**：API、领域规则、关键行为在 `specs/` 下定义。
2. **Spec 为唯一事实来源**：实现与测试以 spec 为准。
3. 详细约定见 [specs/README.md](specs/README.md)。

## 部署（CentOS CVM）

单二进制 + embed 静态资源，可直接在 Linux 上运行。交叉编译、上传、systemd 配置见 **[deploy/README.md](deploy/README.md)**。

## 开发

```bash
# 构建
go build ./...

# 测试
go test ./...
```

## License

未指定；如需开源可自行添加 LICENSE 文件。
