# 配置结构（v2）

本文档按“作用域 + modifier 类型”解释配置 schema，目标是让用户在不读 Go 代码的前提下能写出正确配置。

配置类型定义在 `config/types.go`。

## 顶层结构

- `log`
    - `dir`: 日志目录
    - `level`: 日志级别
    - `fileFormat`: 日志文件名格式

- `server`
    - `address`: 监听地址，例如 `:3000`
    - `tls.cert`/`tls.key`: TLS 证书（可选）

- `modifiers`（可选）
    - 全局 modifiers 列表（作用于所有命中的 action）

- `actions`
    - action 列表（按顺序扫描）

## Action

每个 action 结构：

- `route`: regexp（字符串），用于匹配 `ctx.Path()`
- `type`: `string` 或 `file`
- modifiers（action scope，直接写在 action 下，不再有 `globalmodifier:` 包裹）
- payload：
    - `string`: string payload
    - `file`: file payload

## Modifiers（schema）

所有 modifiers 都通过 `GlobalModifier` 这一套 schema 表达（global/action/payload 三个 scope 都是同一套字段）。

### actionModifierResponseHeader

追加响应头。

示例：

```yaml
actionModifierResponseHeader:
  header:
    Cache-Control:
      - public, max-age=60
```

语义：

- key/value 会通过 fiber `Append` 添加（同名 header 可能追加多个值）

### actionModifierStatus

设置 HTTP status。

```yaml
actionModifierStatus:
  status: 200
```

有效范围：100-599。

未配置时：

- 通常由框架默认决定（大多数情况下为 200）。

### actionModifierAuth

对 request 做鉴权校验。

- `header`: map[string][]string，每个 header 允许多个 regexp
- `query`: map[string]string
- `fallback`:
    - `type: next|jump`
    - `jumpTo`: action index（仅 `jump` 需要）

要求：

- 只要配置了 `actionModifierAuth`，就必须显式配置 `fallback`（否则启动时报错）。

示例：

```yaml
actionModifierAuth:
  header:
    X-Token:
      - ^dev$
  fallback:
    type: jump
    jumpTo: 4
```

语义：

- mismatch -> 触发 fallback
- `next`: `ctx.Next()`（继续 fiber router 链路）
- `jump`: 跳转到 action index（jump target 会按自身配置正常执行，包括其 auth/modifiers）

### actionVersionModifier

驱动版本占位符替换。

```yaml
actionVersionModifier:
  placeholder: ${VERSION}
```

语义：

- placeholder 替换发生在响应头阶段（当前实现针对 response headers）

## Not found（约定）

当所有 actions 都未命中时：router 会无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）。

建议：

- 把最后一个 action 的 `route` 设为空字符串（jump-only），避免它在正常扫描中被误匹配
- 在最后一个 action 上显式配置 `actionModifierStatus: 404`
