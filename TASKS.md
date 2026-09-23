# 并行开发模块

MVP 已打通 server、client、规则引擎、Compose 和测试。多会话开发时，每个会话只领取一个模块，尽量避免修改同一文件。

## A. 指纹规则扩展

主要文件：`rules/fingerprints.json`、`internal/engine/engine_test.go`

- 增加更多 SSH、HTTP Server、数据库和中间件规则；
- 加入 MariaDB、PostgreSQL、MongoDB、SMTP、Telnet、TLS 等；
- 使用表驱动测试验证隐藏数据泛化；
- 避免只依赖端口进行识别。

## B. 规则引擎增强

主要文件：`internal/engine/`

- 支持多个版本捕获组的回退，例如 Jetty 括号版本；
- 将命中特征、端口、产品、版本拆为可解释的评分项；
- 检查重复规则、非法置信度和命名捕获组；
- 添加基准测试和正则输入长度保护。

## C. API 健壮性

主要文件：`internal/api/`

- 精确区分 400、405、413；
- 增加批次数量和单 Banner 长度限制；
- request ID、结构化访问日志和 panic recovery；
- 扩展 httptest 覆盖。

## D. Client 输入兼容

主要文件：`cmd/client/`

- 支持题面中非标准的 `\\xNN` 表达形式；
- 支持 stdin、输出文件和流式错误信息；
- 增加 client 单元测试；
- 保持发送给 server 的内容为标准 JSON。

## E. Docker 与交付验证

主要文件：`Dockerfile`、`compose.yaml`、`README.md`

- 固定基础镜像 digest；
- 验证 `docker compose up --build` 完整流程；
- 增加端到端 smoke test；
- 检查镜像用户、只读文件系统、网络和健康状态。

## F. 文档与验收数据

主要文件：`README.md`、`testdata/`

- 补充 API 契约和规则编写说明；
- 扩充合法 JSON 自测数据与预期输出；
- 提供验收命令和常见问题；
- 不改动核心识别实现。
