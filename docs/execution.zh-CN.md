# 执行流与控制流

本文档聚焦运行时行为：action 扫描、执行器、auth jump、notfound。

相关代码：

- `router/handler/actions.go`
- `action/execute/executor.go`
- `action/modifier_auth.go`

## 扫描策略

router 对 `config.Actions` 做顺序扫描：

- 从 index=0 开始
- 每个 index 调用 `execute.Executor.ExecuteAt(i)`
- 返回 `NotMatched` -> 继续
- 返回 `Matched` -> 结束请求（已写响应）

这一策略意味着：

- action 顺序是可观察的行为
- `next`/`jump` 也基于顺序才有意义

## 执行器（execute.Executor）

执行器承担“执行某一个 index 的 action”的细节：

- 正则匹配（route 是否命中）
- handler 选择（string/file）
- payload 校验（type 与 payload 必须匹配，否则返回 error，不允许 panic）
- modifier 构建与注入（由 router 传入 builders）
- handler 执行

执行器不负责：

- 扫描策略
- jump 后的 index 调整
- notfound

## Auth fallback

### next

auth 校验失败时：执行 `ctx.Next()`。

这意味着请求会回到 fiber router 的后续 handler（在本项目中等价于继续向后执行中间件/路由链）。

### jump

auth 校验失败时：抛出 `action.ErrAuthFallbackJump{JumpTo:int}`。

router 捕获后：

- 将扫描 index 设为 `jumpTo`

route 为空的 action：

- 对正常扫描而言等价于“不匹配”。
- 对 jump 而言会被强制执行（跳转不依赖 route 匹配）。

要求：

- jumpTo 必须是有效下标
- 应避免配置导致的 jump 循环

## Not found

扫描结束仍未命中时：

- 无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）
    - 最后一个 action 若返回 error：直接向上返回，交给 fiber error handler 处理

注意：这个设计意味着 notfound 是“可编程的兜底 action”。为了安全，强烈建议：

- 最后一个 action 明确设置 `actionModifierStatus: 404`
- 最后一个 action 做好鉴权（`actionModifierAuth`），避免泄露文件或敏感信息
