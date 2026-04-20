# 架构说明

本文档描述 CatSync v2 的模块划分、依赖边界和主要数据流。

目标是让维护者知道：

- 哪些包负责什么
- 包之间通过什么数据结构交互
- 哪些逻辑应该放在哪一层

## 包与职责

### `config`

- 定义配置结构（`config/types.go`）
- 负责加载/保存/合并/校验（`config/config.go`, `config/loader/*`）
- 校验方法已全部内聚为 `*Config` 的私有绑定，避免全局散落野生函数。
- `config.GetDefaultConfig()` 提供默认配置（`config/default.go`）

依赖方向：

- `config` 不能依赖 `router`/`action` 的运行时代码。
- `config` 可以依赖 `internal/util` 的通用反射/校验工具。

### `action`

- 定义 action handler 的运行时接口与上下文（`action/handler.go`）
- 提供基础 handler：`string` / `file` / `server` / `reload`（`action/action.go`）
- 提供 modifier 系统（`action/modifier.go` + 各类 `action/modifier_*.go`）
- 依赖于内联方法与早返回，高度压平循环嵌套以保证零分配与性能。
- 通过 `HandlerRegistry` 将 `config.ActionType` 映射到运行时 handler（`action/action.go`, `action/registry.go`）

依赖方向：

- `action` 可以依赖 `config`（因为 `ProcessData` 里包含 action/payload）。
- `action` 不应该依赖 `router`（避免反向依赖）。

### `action/execute`

这是系统的**运行时心脏（Executor）**，将配置态转为运行态：

- 通过 `Executor.Build` 对配置中的所有 Modifier 层级进行洋葱模型（Decorator）**预构建**。
- 提供极速的 `MatchRoute` 机制，其中精确匹配路由被转换为 `map` 的 **O(1)** 查询。
- 提供完整的请求分发引擎（`Executor.Dispatch`），内聚跳转链（Jump）跟踪与防死锁调度逻辑。
- 返回 `Result` 及其状态（`StatusMatched`, `StatusJump`, `StatusNotMatched`）。

### `router`

    - 封装 Fiber HTTP Server 中间件层（Recover, Compress, CORS 等）。
    - 向 `execute.Executor` 注册请求上下文（Context）。
- 核心路由文件 `actions.go` 已完全扁平化为纯配置态入口（将请求统交由 `Dispatch` 处理）。

### `framework`

- 提供 Fiber/HTTP server 的启动封装。
- 绑定 `router.Init()`。
- 原生实现与配置 Let's Encrypt ACME (DNS-01/HTTP-01) 证书。

## 数据流（请求路径）

**A. 初始化态 (Init / Reload)**
1. `cmd/main.go` -> `config.New()` 加载所有配置。
2. 执行引擎预构建：`execute.New().WithConfig().Build()`。预分配所有的 Action Entry 和 Modifier 链，建立 O(1) 路由。

    **B. 运行态 (Request)**
    1. `framework.NewFiber` 构造 fiber app 并在 `router.Init` 注册 middleware 和 `handler.Actions`。
    2. HTTP 请求进入 `router/handler/actions.go`。
    3. `exec.Dispatch(rc)`：
    - 通过 O(1) Map 或 Regex Scanner 寻找命中的 Action Index。
    - 若存在，执行这个 Action `executeOne`。
    - 若触发 Auth Fallback 跳转（`JumpTo`），通过高效的位图 Bitset 状态机跳转去目标 Index，防环。
    - 执行至对应 Handler（例如 `FileHandler`），由 Base Handler 输出响应 `SendFile` / `SendString`。
    - 如果没有任何命中 -> 无条件执行最后一个 Action 作为 Notfound Handler。

## 依赖边界建议

- HTTP 层级的 Middleware 等全局拦截逻辑放 `router`。
- action-level 的原子执行能力放 `action`。
- 复杂的“条件判断、跳转引擎、路由树”放 `action/execute`。
- 不要让 `config` 里出现运行时对象（如 Modifier 实例、Fiber 等）。
