# 设计: Modifiers 优先的 Action 与执行流

本文档描述 CatSync v2 的配置模型与请求执行架构。

核心变化是从“零散的 action 顶层字段”（例如 `auth`、`responseHeader`、`status`）迁移到统一的 modifiers-first
模型，并引入一层可复用的执行器来承载 auth fallback（包含 `jumpTo`）等控制流。

同时说明 router 级别的 not found 行为，以及为什么移除专门的 `error/notfound` action type。

## 目标

- 用同一种机制表达 action 的横切能力（header、status、auth、version placeholder 等）。
- 支持多级作用域（global / action / payload）配置 modifiers，避免 schema 继续膨胀。
- 明确定义并实现 auth fallback 行为（`next` / `jumpTo`），保证可预测、可调试。
- 执行流尽量直观，不把控制流藏在 router 的杂糅循环里。
- 默认配置能演示完整功能，且尽量开箱即用。

## 非目标

- 不是通用策略引擎；modifier 设计为小而可组合。
- 不是 workflow 系统；控制流只限定在 auth fallback 这一类场景。
- 本轮为破坏性更新，不承诺严格向后兼容。

## 术语

- Action: `config.Actions` 中的一条路由条目，包含 `route` regexp 与 `type`。
- Payload: action 的 `string` 或 `file` 配置块。
- Modifier（运行时）: `action.Modifier`，用于包装/增强 handler。
- Modifier 配置（schema）: `config/types.go` 中的配置类型，用于构建运行时 modifier。

## 配置模型

### 三个作用域

Modifiers 分三层存在：

1. Global（全局）: `config.Config.Modifiers`
    - 对所有命中的 action 生效。
    - 用途: 默认响应头、全局 auth 规则、默认 status 等。

2. Action（单条 action）: `config.Action` 内嵌 `GlobalModifier`
    - 仅对当前 action 生效。
    - 用途: 该路由的专属 header/status/auth。

3. Payload（数据块）: `config.ActionStringData` / `config.ActionFileData` 内嵌 `GlobalModifier`
    - 对具体的 payload 生效。
    - 用途: 更贴近 payload 的配置（例如 string 的 content-type、file 的 cache-control）。

router 构建运行时 modifiers 的顺序为：

1. global modifiers
2. action modifiers
3. payload modifiers
4. type-specific modifiers（目前仅: version placeholder）

这保证了组合行为可预测：先应用默认值，再应用更具体的配置。

### 移除 legacy 顶层字段

以下字段已从 `config.Action` 移除：

- `auth`（改为 `actionModifierAuth`）
- `responseHeader`（改为 `actionModifierResponseHeader.header`）
- `status`（改为 `actionModifierStatus.status`）

原因：

- 这些字段本质是“横切能力”，放在顶层会让 schema 随功能增长不断加字段。
- 统一为 modifiers 后，扩展点集中在 modifier schema，不需要再开新的 action 字段。

### Not found 行为

not found 不是“某个 action 的行为”，而是“所有 action 都没命中”的结果。

v2 采用一个确定的约定：

- 当扫描结束仍未命中时，router 无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）。

这样实现更确定：router 扫完 actions 后统一处理；并且 notfound 也可以用普通 modifiers（status/headers/auth）表达。

## 运行时模型

### Handler pipeline

请求执行过程：

1. router 遍历 actions，基于 `route` regexp 选择匹配项。
2. 根据 `type` 从 `action.HandlerRegistry` 取出 base handler（`string` / `file`）。
3. 将 handler 包装为 `action.ModifiableHandler`，依次注入运行时 modifiers。
4. 构造 `action.ProcessData` 并执行 handler。

运行时 modifiers 由单个 `action.ModifiableHandler` 按注册顺序执行 `Before`，执行 base handler 后按逆序执行 `After`。

### Version placeholder

version placeholder 通过 `action.PlaceholderModifier` 实现，配置由 `actionVersionModifier.placeholder` 驱动。

placeholder modifier 会扫描响应头 key/value，将占位符替换为真实版本号。

### Auth modifier 与 fallback 控制流

auth 校验由 `action.AuthModifier` 实现，支持：

- header 校验（每个 header 可配置多个 regexp）
- query 校验

当校验失败时，fallback 由 `ActionModifierAuthFallback` 控制：

- `next`: 走 `ctx.Next()`，继续 fiber 的路由链
- `jump`: 跳到 `config.Actions[jumpTo]`

#### jumpTo 语义

`jumpTo` 是 actions 的下标（index）。

`jump` 的语义是“改变执行位置”，而不是“绕过鉴权”。
因此跳转目标是否需要鉴权，完全由目标 action 自身的 modifiers 决定。

实现方式：

- auth modifier 在 mismatch 时返回 `action.ErrAuthFallbackJump{JumpTo:int}`。
- 执行循环捕获该 error，设置下一次执行下标为 `jumpTo`。

约束与建议：

- 配置 `jump` 时，应确保跳转链路不会形成循环。
- 推荐 jump 到一个明确的 fallback action（通常不配置 auth，直接返回 401/403 或提示信息）。

## 执行架构

### 可复用执行器（公开 API）

执行单条 action 的逻辑封装为可复用组件：

- package: `action/execute`
- type: `execute.Executor`

使用 With 风格注入依赖：

- `WithConfig(*config.Config)`
- `WithBuilders(execute.Builders{...})` 或 `WithGlobalBuilder/WithActionBuilder/WithPayloadBuilder`
- `WithContext(*params.Ctx, *fiber.Ctx)`
- `WithSkipRouteCheck(bool)`

执行入口：

- `ExecuteAt(index int) (Result, error)`

`Result` 只表达 router 需要的控制信息：

- `NotMatched`: 当前 index 不匹配该请求，router 继续扫描
- `Matched`: 命中并成功执行
- `JumpTo`: auth fallback 请求跳转到下标

### Router 的职责边界

router 只负责调度：

- 遍历 index 并调用 `Executor.ExecuteAt(i)`
- `NotMatched` 继续扫描
- `JumpTo` 改写扫描下标，并标记 jump 目标
- 扫描结束仍未命中，则无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）

这样控制流集中在 router，执行细节集中在 executor。

## 默认配置与示例约束

默认配置应覆盖：

- global response headers
- status modifier
- auth modifier（同时演示 `next` 和 `jump`）
- version placeholder
- file action
- notfound handler（last action）

并尽量避免需要用户额外准备文件。

### File action 的健壮性

为了让默认配置开箱即用，file action 对 missing/empty 文件做兜底：

- 文件不存在：创建父目录与文件，写入 `CatSync!\n`
- 文件存在但为空：写入 `CatSync!\n`

这是一种偏“演示/易用”的策略。如果未来要更严格，可以把文件初始化移到启动阶段而不是 handler 内。

## 已知限制与后续方向

- “覆盖语义”尚未在所有 modifiers 上统一成严格规则。目前是顺序叠加。
  对 response headers 来说等价于 append；如果希望 override，需要为每个 modifier 定义 resolution 规则。

- `WithSkipAuth` 会修改 executor 的状态。当前每个请求串行使用同一个 executor 实例，语义安全。
  若未来出现并发复用，应改为 copy-on-write（With 返回浅拷贝）。
