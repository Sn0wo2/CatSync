# 执行流与控制流

本文档聚焦运行时行为：action 扫描、执行器、auth jump、notfound。

相关代码：

- `router/handler/actions.go`
- `action/execute/executor.go`
- `action/modifier_auth.go`

## 扫描策略

router 对 `config.Actions` 交由 `execute.Executor.Dispatch` 统一调度：

- 内部对请求进行路由匹配：精确路由使用 Map 实现 **O(1)** 极速查找，正则路由顺序匹配预编译的正则表达式。
- 匹配成功并执行完毕 -> 返回 `Matched` -> 结束请求（已写响应）。
- 执行过程返回 `NotMatched` 或不匹配 -> 继续寻找下一个匹配的 Action。

这一策略意味着：

- action 的声明顺序是可观察的匹配行为依据。
- `next`/`jump` 也基于此顺序才有意义。

## 执行器（execute.Executor）

执行器承担了预编译和“执行某一个 index 的 action”的所有细节，极大降低了请求时的运行时开销：

- **预编译机制 (Build)**：在系统启动或配置 Reload 时，Executor 会一次性预构建所有的 Modifier 层级和 Handler 实例（零分配），并将路由分类缓存。
- **路由匹配**：提供高性能的路径判断。
- **跳转控制**：在内部通过 Bitset 处理安全的 Jump 跳转链，防环死锁。
- **Handler 执行**：将请求流转入具体的业务 Handler。

执行器不负责：

- Notfound 的默认行为渲染（交由 `ServerHandler` 等叶子节点或用户显式配置的最终 Action 兜底）。

## Auth fallback

### next

auth 校验失败时：触发 `AuthFallbackNext`。

这意味着执行器将当前 Action 视为 `NotMatched`，继续按照顺序向后扫描其余可能匹配的 Action 路由。

### jump

auth 校验失败时：抛出 `action.ErrAuthFallbackJump{JumpTo:int}`。

`Executor.Dispatch` 捕获后：

- 立即跳转至目标 `jumpTo` 对应的 Action Index 执行。
- 内部限制最大跳转深度（`maxJumpDepth = 16`），同时使用位图检测并防御循环跳转（Loop Detected）。

route 为空的 action：

- 对正常的 HTTP 路径扫描而言等价于“不匹配”，永远不会被直接访问。
- 对 jump 而言会被强制执行（跳转行为跳过路由匹配检查），非常适合用来做纯逻辑或兜底页面的“隐藏路由”。

要求：

- jumpTo 必须是有效下标。
- 应避免配置导致死循环。

## Not found

扫描结束仍未命中时：

- 无条件执行配置中的最后一个 action 作为 notfound handler（此时跳过 route 匹配强制执行）。
    - 若最后一个 action 返回 error：向上抛出并交给 Fiber 全局 Error Handler 统一处理（记录 TraceID 和 Stack）。

注意：这个设计意味着 notfound 是“可编程的兜底 action”。为了安全，强烈建议：

- 最后一个 action 明确设置 `actionModifierStatus: 404`。
- 最后一个 action 做好鉴权（`actionModifierAuth`），避免泄露文件或敏感信息。
