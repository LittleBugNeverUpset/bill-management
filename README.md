# 开发计划

基础层→核心层→业务层→扩展层」逐步推进

阶段 1：基础环境与核心工具封装（2-3 天）
目标：搭建项目骨架，封装通用工具，验证基础环境连通性，这是所有业务的基础。

| 步骤 | 开发内容 | 测试方式 |
|:----:|------|------|
| **1.1** | **初始化项目**<br>- 创建 go.mod，引入核心依赖（gin、gorm、go-redis、zap、jwt、viper、gin-swagger）<br>- 按既定结构创建所有空目录 | 执行 `go mod tidy`，确认依赖下载成功 |
| **1.2** | **配置管理（configs/ + pkg/）**<br>- 编写 `configs/config.yaml`（MySQL、Redis、JWT、日志配置）<br>- 封装 viper 读取配置（`pkg/config/config.go`） | 编写单元测试，验证配置能正确读取（如打印 MySQL 地址） |
| **1.3** | **日志工具封装（pkg/logger/）**<br>- 封装 zap 日志，支持控制台 / 文件输出，区分不同日志级别 | 测试日志能否正常写入文件，控制台输出格式是否正确 |
| **1.4** | **数据库 / 缓存客户端封装（pkg/）**<br>- 封装 MySQL 连接（`pkg/mysql/mysql.go`），验证连接池<br>- 封装 Redis 客户端（`pkg/redis/redis.go`），验证 ping 连通性 | 单元测试：测试 MySQL/Redis 能否正常连接，失败时抛出明确错误 |
| **1.5** | **通用工具封装（pkg/）**<br>- 统一响应格式（`pkg/response/response.go`：成功 / 失败 / 分页响应）<br>- 密码加密（`pkg/crypto/crypto.go`：bcrypt 加密 / 验证） | 单元测试：验证密码加密后能正确验证，响应格式输出符合预期 |


阶段 2：核心鉴权模块（1-2 天）
目标：完成 JWT 鉴权核心逻辑，能独立测试「生成 token→验证 token→注销 token」全流程。

| 步骤 | 开发内容 | 测试方式 |
|------|----------|----------|
|**2.1**| JWT 工具封装（pkg/jwt/jwt.go）<br>- 实现 token 生成、验证、解析 claims<br>- 实现 token 黑名单（Redis 存储） | 单元测试：<br>1. 生成 token 后能正确解析出 UserID<br>2. 注销 token 后，验证接口返回「token 已失效」 |
| **2.2**| 鉴权中间件（internal/middleware/jwt.go）<br>- 从 Header 获取 token，验证有效性，将 UserID 存入 Gin Context | 集成测试：<br>1. 无 token 请求 → 返回 401<br>2. 无效 token → 返回 401<br>3. 有效 token → 能从 Context 获取 UserID |
| **2.3** | 日志中间件（internal/middleware/logger.go）<br>- 集成 zap 日志，记录请求方法、路径、状态码、耗时 | 测试任意请求，查看日志文件是否记录完整请求信息 |