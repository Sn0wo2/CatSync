# 架构说明

本文档描述 CatSync v2 的模块划分、依赖边界和主要数据流。

目标不是画一张“看起来很大”的图，而是让维护者知道：

- 哪些包负责什么
- 包之间通过什么数据结构交互
- 哪些逻辑应该放在哪一层

## 包与职责

### `config`

- 定义配置结构（`config/types.go`）
- 负责加载/保存/合并/校验（`config/config.go`, `config/loader/*`）
- `config.GetDefaultConfig()` 提供默认配置（`config/default.go`）

依赖方向：

- `config` 不能依赖 `router`/`action` 的运行时代码
- `config` 可以依赖 `internal/util` 的通用反射/校验工具

### `action`

- 定义 action handler 的运行时接口与上下文（`action/handler.go`）
- 提供基础 handler：`string` / `file`（`action/action.go`）
- 提供 modifier 系统（`action/modifier.go` + 各类 `action/modifier_*.go`）
- 通过 `HandlerRegistry` 将 `config.ActionType` 映射到运行时 handler（`action/action.go`, `action/registry.go`）

依赖方向：

- `action` 可以依赖 `config`（因为 `ProcessData` 里包含 action/payload）
- `action` 不应该依赖 `router`（避免反向依赖）

### `action/execute`

这是 router 可复用的一层：

- 封装“执行某一个 action index”的逻辑
- 返回 `Result` 供 router 做调度（`NotMatched` / `JumpTo` 等）
- 不负责扫描策略（扫描策略是 router 的职责）

### `router`

- 定义请求到 action 的扫描/调度策略
- 负责:
    - 正则匹配 `route`
    - 执行器调度（循环、处理 jump、notfound）
    - 中间件（compress/cors 等）

router 不应该把“执行一个 action 的细节”写成大块逻辑；那部分在 `action/execute`。

### `framework`

- 提供 Fiber/HTTP server 的启动封装
- 绑定 `router.Init()`

## 数据流（请求路径）

1. `framework.NewFiber` 构造 fiber app
2. `router.Init` 注册 middleware 和 `handler.Actions`
3. 请求进入 `router/handler/actions.go`：
    - 构造 `execute.Executor`（per-request）
    - 遍历 `cfg.Actions`，按 index 执行
    - `NotMatched` -> 继续
    - `JumpTo` -> 修改 index
    - 扫描结束未命中 -> 无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）
4. `execute.Executor.ExecuteAt`：
    - 校验 payload
    - 构建 handler + modifiers
    - 构造 `action.ProcessData` 调用 handler
5. `action` handler（string/file）输出响应

## 依赖边界建议

- router-level 的策略（扫描、notfound）放 `router`
- action-level 的能力（auth/header/status/version）放 `action` + `config` schema
- “执行一个 action”放 `action/execute`
- 不要让 `config` 里出现运行时对象（modifier 实例、handler 等）
