# Fingerprint Service

Fingerprint Service 是一个轻量、无状态的网络服务指纹识别工具。客户端从本地 JSON 文件读取批量数据，调用服务端的 HTTP API；服务端根据 YAML 规则对 Banner 进行匹配、提取产品与版本信息，并按输入顺序返回识别结果。

## 设计目标

- **Banner 优先**：Banner 是主要识别依据，端口只参与辅助判断，避免非标准端口造成误判。
- **规则与代码分离**：协议、产品、版本和操作系统提示统一维护在 `rules/fingerprints.yaml` 中。
- **批量容错**：单条数据无法识别时返回 `unknown`，不会中断整批请求。
- **顺序稳定**：响应结果与输入数据保持相同顺序，便于调用方关联。
- **服务无状态**：不依赖数据库或本地会话，便于扩容、容器化和离线测试。

## 系统结构

```text
本地 JSON 文件
      │
      ▼
fingerprint-client
  - 读取与校验输入
  - POST /fingerprint
  - 格式化输出
      │
      │ Docker 内部网络
      ▼
fingerprint-server
  - HTTP API
  - 批量识别
  - 规则匹配与评分
      │
      ▼
fingerprints.yaml
  - 协议规则
  - 产品规则
  - 版本提取
  - OS 提示
```

## 目录结构

```text
.
├── cmd/
│   ├── server/              # 服务端入口
│   └── client/              # 命令行客户端入口
├── internal/
│   ├── api/                 # HTTP Handler 与中间件
│   ├── engine/              # 匹配、评分及识别流程
│   ├── model/               # 请求、响应与指纹模型
│   └── rules/               # YAML 规则加载与校验
├── rules/
│   └── fingerprints.yaml    # 指纹规则库
├── testdata/                # 示例输入与预期输出
├── Dockerfile
├── compose.yaml
└── go.mod
```

## API 约定

### 健康检查

```http
GET /healthz
```

成功时返回 HTTP `200 OK`。

### 批量指纹识别

```http
POST /fingerprint
Content-Type: application/json
```

请求示例：

```json
{
  "services": [
    {
      "host": "192.0.2.10",
      "port": 2222,
      "banner": "SSH-2.0-OpenSSH_9.6p1 Ubuntu-3ubuntu13"
    },
    {
      "host": "192.0.2.11",
      "port": 9000,
      "banner": "unrecognized service"
    }
  ]
}
```

响应示例：

```json
{
  "results": [
    {
      "host": "192.0.2.10",
      "port": 2222,
      "protocol": "ssh",
      "product": "OpenSSH",
      "version": "9.6p1",
      "os": "Ubuntu",
      "confidence": 0.98,
      "status": "identified"
    },
    {
      "host": "192.0.2.11",
      "port": 9000,
      "protocol": "unknown",
      "product": "unknown",
      "version": "",
      "os": "unknown",
      "confidence": 0,
      "status": "unknown"
    }
  ]
}
```

服务端应为无效的整体请求返回 `4xx`；对于格式有效但无法识别的单条服务，应在对应位置返回 `unknown`。

## 本地运行

环境要求：Go 1.24 或更高版本。

```bash
go run ./cmd/server
```

默认约定：

- 监听地址：`:8080`
- 规则文件：`rules/fingerprints.yaml`
- `SERVER_ADDR` 可覆盖监听地址。
- `RULES_FILE` 可覆盖规则文件路径。

调用服务：

```bash
go run ./cmd/client \
  --input testdata/input.json \
  --server http://localhost:8080 \
  --output -
```

其中 `--output -` 表示将格式化后的 JSON 输出到标准输出。

## Docker 运行

启动服务端：

```bash
docker compose up --build -d server
```

服务默认暴露在 `http://localhost:8080`。如需修改宿主机端口：

```bash
FINGERPRINT_PORT=18080 docker compose up --build -d server
```

通过 Docker 内部网络运行客户端，并读取 `testdata/input.json`：

```bash
docker compose --profile client run --rm client
```

查看服务日志或停止容器：

```bash
docker compose logs -f server
docker compose down
```

也可以分别构建两个镜像目标：

```bash
docker build --target server -t fingerprint-server .
docker build --target client -t fingerprint-client .
```

## 规则维护

规则文件应至少能够描述以下信息：

- 协议或产品的 Banner 匹配表达式；
- 可选的常见端口，用于辅助加权而不是作为唯一依据；
- 产品名称与协议名称；
- 版本号捕获规则；
- 操作系统提示；
- 规则权重或置信度贡献。

新增产品时，优先增加或调整 YAML 规则。只有在现有规则模型无法表达识别逻辑时，才修改 Go 匹配引擎。

## 测试

```bash
go test ./...
```

`testdata/input.json` 用于保存稳定的批量输入，`testdata/expected.json` 用于保存对应的预期结果。测试应覆盖非标准端口、未知 Banner、版本提取失败和批量顺序保持等场景。
