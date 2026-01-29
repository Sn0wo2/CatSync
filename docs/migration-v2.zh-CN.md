# v2 迁移指南（破坏性更新）

本文档面向从 v1 配置迁移到 v2 的用户。

v2 的配置模型发生了结构性变化：从“在 action 顶层塞功能字段”改为 modifiers-first（同一套 modifier schema 在
global/action/payload 三个 scope 复用）。
同时路由执行架构也改变了：auth/headers/status 等不再由 router 直接处理，而是由 runtime modifiers 统一处理。

这是一份偏实战的迁移手册：

- 列出主要 break changes
- 给出 v1 -> v2 的等价写法
- 提醒行为差异与坑

相关文档：

- 快速开始：[`docs/quick-start.zh-CN.md`](quick-start.zh-CN.md)
- 详细配置指南：[`docs/config-guide.zh-CN.md`](config-guide.zh-CN.md)
- 执行流说明：[`docs/execution.zh-CN.md`](execution.zh-CN.md)

## 一句话总结

- v1: `Action` 顶层字段驱动行为（`auth/responseHeader/status/server.header` 等），router 里直接做校验与 header 写入。
- v2: 统一使用 modifiers（`actionModifierAuth/actionModifierResponseHeader/actionModifierStatus/...`），router 只负责扫描与调度。

v2 的意义（给升级决策者/维护者）：

- 配置表达能力更统一：新能力以 modifier 形式加入，不需要再给 `Action` 增加新的顶层字段。
- 执行逻辑模块化：auth/header/status/version 等从 router 拆到 `action` modifiers，减少 router 复杂度。
- 可组合：同一套 modifier schema 在 global/action/payload 复用，减少重复配置。
- 性能更稳定：regexp 编译缓存（`internal/util.GetCompiledRegexp`），避免每个请求重复 compile。
- 行为更可测试：modifiers 与执行器可以被单独测试/审查。

## Break changes 总览

1. `Action` 顶层字段移除：`auth` / `responseHeader` / `status`
2. `server.header` 移除：统一使用 response header modifier 设置 `Server`
3. `error` / `notfound` action type 移除：notfound 由“最后一个 action”承担
4. file action 字段变更：`dontDetectContentType` -> `dontSetContentType`
5. YAML 结构变更：action/payload 的 modifiers 直接 inline（不再出现 `globalmodifier:` 外壳）
6. 路由匹配实现变更：从 v1 的直接匹配/存在 bug 的 regexp 使用，到 v2 的 regexp + 缓存
7. auth fallback 支持 `next/jump`，`jumpTo` 为 action index；jump 不再绕过 auth
8. action.route 允许为空：空 route 的 action 不参与正常匹配，只能作为 jump 目标
9. router 不再内联处理 auth/header/status，而是通过 modifiers 统一实现

## 字段映射速查表（v1 -> v2）

| v1                                    | v2                                                                    | 说明                            |
|---------------------------------------|-----------------------------------------------------------------------|-------------------------------|
| `actions[].auth`                      | `actions[].actionModifierAuth`                                        | 行为从 router 下沉到 modifier       |
| `actions[].responseHeader`            | `actions[].actionModifierResponseHeader.header` 或 global/payload 同名字段 | v2 仍是 append 语义               |
| `actions[].status`                    | `actionModifierStatus.status`                                         | 建议在 action 或 payload scope 设置 |
| `server.header`                       | `modifiers[].actionModifierResponseHeader.header.Server`              | 推荐放 global                    |
| `string.actionVersionModifier`（v1 内嵌） | `actionVersionModifier.placeholder`（v2 指针字段，写法一致）                     | v2 主要用于响应头 placeholder        |
| `file.dontDetectContentType`          | `file.dontSetContentType`                                             | 语义相反的命名，迁移时注意                 |

## 迁移步骤（建议顺序）

1. 先把旧配置文件备份一份
2. 按 v2 schema 先跑起来（可以先用 v2 默认 config，对照改）
3. 迁移 `server.header` -> global response header modifier
4. 逐条迁移 actions：
    - 把 `auth/responseHeader/status` 挪到对应 modifier
    - 确保 `type` 与 payload 块匹配
5. 补上 notfound handler（把最后一个 action 配成 404 兜底）
6. 检查 file action 字段名变更与路径约束
7. 用 curl 验证关键路由与 auth 行为

## Router 行为变更（v1 -> v2）

这一节专门说明 router 的 break changes，因为它直接影响 route 写法、action 顺序语义和性能。

### v1 路由的典型问题（以 tag 版本为参考）

在 v1.6.x 和 v2 beta.5/beta.6 的旧实现中，router 往往承担了过多逻辑：

- route 匹配、auth 校验、response header 写入都在 router 内完成
- 正则频繁编译，缺少缓存

并且在部分版本里 route 匹配存在明显 bug：

- `regexp.Compile(act.Route)` 后使用 `re.MatchString(act.Route)`（匹配的是 route 自身而不是 request path）

这类问题会导致：

- 看起来“写了 route 但不生效/总是 Next”
- 维护成本高：任何横切能力的变更都要动 router

### v2 的策略

v2 的 router 只做两件事：

1. 扫描与调度（包含 jump/notfound）
2. 把“执行单个 action index”的细节委托给 `execute.Executor`

对应的能力下沉到 modifiers：

- auth -> `actionModifierAuth`
- response headers -> `actionModifierResponseHeader`
- status -> `actionModifierStatus`

### route 写法建议

建议把 `route` 当作“完整路径匹配的 regexp”来写，并显式使用 `^...$`：

```yaml
route: ^/public/hello\\.txt$
```

原因：

- action 顺序在 v2 是配置语义的一部分（next/jump 都依赖顺序）
- 使用锚点可以减少误匹配，避免某个 action 意外吞掉其它路径

## 顶层与 Action 字段迁移

### 1) server.header 移除

v1:

```yaml
server:
  header: CatSync
```

v2: 使用 response header modifier 设置 `Server`（推荐放全局）：

```yaml
modifiers:
  - actionModifierResponseHeader:
      header:
        Server:
          - CatSync
```

注意：v2 仍使用 `Append` 追加 header，同名 key 多次出现会追加多个值。

### 2) Action 顶层 auth/responseHeader/status 移除

v1 的 action 往往长这样：

```yaml
actions:
  - route: ^/secure$
    type: string
    status: 200
    responseHeader:
      Content-Type:
        - text/plain; charset=utf-8
    auth:
      header:
        X-Token:
          - ^dev$
    string:
      content: ok
```

v2 迁移要点：把这些字段改成 modifiers（scope 可选：global/action/payload）。

等价写法（action scope）：

```yaml
actions:
  - route: ^/secure$
    type: string
    actionModifierStatus:
      status: 200
    actionModifierResponseHeader:
      header:
        Content-Type:
          - text/plain; charset=utf-8
    actionModifierAuth:
      header:
        X-Token:
          - ^dev$
      fallback:
        type: next
    string:
      content: ok
```

也可以把 `Content-Type` 放到 payload scope（更贴近 string/file 本身）：

```yaml
actions:
  - route: ^/secure$
    type: string
    actionModifierAuth:
      header:
        X-Token:
          - ^dev$
      fallback:
        type: next
    string:
      actionModifierStatus:
        status: 200
      actionModifierResponseHeader:
        header:
          Content-Type:
            - text/plain; charset=utf-8
      content: ok
```

## action type 变化

以下 action type 已移除：

- `error`
- `notfound`

原因：

- notfound 属于 router 行为
- error 作为 action type 会让路由/执行边界变得模糊

替代方式：

- 把最后一个 action 当作 notfound handler（建议 `route: ""`）
- 其它错误场景由 handler/modifier 返回 error，并由 fiber error handler 统一处理

notfound 示例：

```yaml
actions:
  # ...其它 actions...

  - route: ""
    type: string
    actionModifierStatus:
      status: 404
    string:
      content: "page not found\n"
```

行为差异：

- v1 的 router 在扫描结束后直接 `return nil`，这类行为在不同部署环境下可能表现不一致。
- v2 推荐显式配置 notfound，避免依赖框架默认。

## auth fallback

v2 的 auth fallback 由 `actionModifierAuth.fallback` 控制：

- `type: next` -> `ctx.Next()`
- `type: jump` + `jumpTo` -> 跳到 action index

### v1 vs v2 的 auth 配置差异

v1（旧模型）通常在 action 顶层写 `auth`：

```yaml
actions:
  - route: ^/secure$
    type: string
    auth:
      header:
        X-Token:
          - ^dev$
    string:
      content: ok
```

v2（新模型）写到 modifier：

```yaml
actions:
  - route: ^/secure$
    type: string
    actionModifierAuth:
      header:
        X-Token:
          - ^dev$
      fallback:
        type: next
    string:
      content: ok
```

### jump 语义（重要变更）

`jumpTo` 是 actions 的下标（index）。

注意：v2 的 `jump` 只改变执行位置，不会自动绕过 auth。
跳转目标是否鉴权完全由目标 action 自己的 modifiers 决定。

因此建议的写法是：

- 让 jump 指向一个明确的 fallback action（通常不配置 auth），负责返回 401/403 或提示信息。
- 避免配置形成 jump 环（例如 A -> B -> A）。当前实现会检测重复 jump 目标并返回错误。

## 默认配置生成行为

- 如果指定了 `CONFIG_PATH` / `DEBUG_CONFIG_PATH` 但文件不存在，程序会创建默认 config 到该路径。

## file action 迁移要点

### 字段名变更

v1:

```yaml
file:
  path: ./data/hello.txt
  dontDetectContentType: true
```

v2:

```yaml
file:
  path: ./data/hello.txt
  dontSetContentType: true
```

### 路径约束

v2 会校验 file path 必须在 `./data` 目录下（防止目录穿越）。

### missing/empty 文件行为

为保证默认示例可用，v2 对 missing/empty 文件会自动写入 `CatSync!\n`。
如果你依赖严格的 “文件不存在就报错” 行为，需要在部署侧保证文件存在且不为空。

## 维护者视角：代码变更整合与边界调整

这一节写给维护者，关注点是“为什么这么改”和“代码边界怎么分”。

### 1) 行为从 router 下沉到 modifiers

v1 的 router 直接做：

- auth 校验
- response header 写入
- status 写入

缺点：

- router 变成大杂烩，任何横切能力变更都要修改 router
- 行为难以复用（例如未来新增入口/框架，无法复用执行逻辑）

v2：

- 这些能力变成 modifiers（`action/modifier_*.go`）
- router 只保留扫描与调度（`router/handler/actions.go`）

### 2) 引入执行器（execute.Executor）

执行器封装“执行某一个 action index”的细节：

- route 匹配
- payload 校验（type 与 payload 必须匹配，避免 nil panic）
- modifiers 注入
- handler 执行

router 负责：

- for-loop 扫描
- jump 控制流
- notfound

### 3) notfound 使用最后一个 action（约定）

notfound 是“扫描结束仍未命中”的结果，属于 router 层。
v2 采用一个简单、确定的约定：

- router 总是无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）
- 最后一个 action 若执行返回 error：会直接向上返回，交给 fiber error handler 处理

安全建议：

- 最后一个 action 建议 `route: ""`，并显式 `status: 404`
- 如果最后一个 action 是 file 或不做鉴权，可能导致敏感内容泄露

### 4) action 顺序成为配置语义

v2 的 `next/jumpTo` 都依赖 action index：

- `jumpTo` 是 index
- `next` 会继续向后扫描

维护建议：

- 将兜底 action 放在列表末尾
- 明确写出 `^...$` 的 route，减少误匹配

## 迁移注意事项（常见坑）

### 1) type 与 payload 必须匹配

- `type: string` 必须提供 `string:`
- `type: file` 必须提供 `file:`

v2 执行器会对不匹配的情况直接返回 error（避免 nil pointer panic）。

### 2) route 现在通常建议写 regexp

建议显式写 `^...$`，避免误匹配：

```yaml
route: ^/$
```

### 3) response header 是 append

`actionModifierResponseHeader` 使用追加语义，同名 header 多次配置会追加多个值。
如果你之前依赖“覆盖”，需要在配置上避免重复 key，或后续引入 replace 语义。

### 4) file action 路径约束

file action 的 `path` 必须在 `./data` 目录下（运行时会校验）。

### 5) action 顺序变成配置语义

v2 的扫描与 jump 都依赖 index 顺序：

- `jumpTo` 是 index，不是 route
- `next` 会继续扫描后续 action

迁移时建议把 actions 按“从具体到兜底”的顺序排列，并把 fallback action 放在末尾。

### 7) route 为空的 action

v2 允许 `route: ""`。

- 这类 action 不参与路径匹配扫描
- 只能通过 `jumpTo` 进入执行

建议：

- jump-only action 的 index 在迁移/维护过程中要特别小心（index 变动会改变 jump 行为）

### 6) auth regexp 现在会被缓存

v1 在 router 内部频繁 `regexp.Compile`。
v2 使用 `internal/util.GetCompiledRegexp` 做缓存与复用（性能更稳定）。
迁移时注意：regexp 语法错误会在运行时以更明确的 error/warn 暴露。

## 完整示例（v1 -> v2）

下面给一份典型 v1 配置与其 v2 等价写法，便于整体对照。

### v1 示例

```yaml
log:
  dir: ./logs
  level: debug
  fileFormat: 2006-01-02.log

server:
  address: :3000
  header: CatSync

actions:
  - route: ^/$
    type: string
    status: 200
    responseHeader:
      Content-Type:
        - text/plain; charset=utf-8
    string:
      content: Hello

  - route: ^/secure$
    type: string
    auth:
      header:
        X-Token:
          - ^dev$
    string:
      content: ok
```

### v2 示例

```yaml
log:
  dir: ./logs
  level: debug
  fileFormat: 2006-01-02.log

server:
  address: :3000
  tls:
    cert: ""
    key: ""

modifiers:
  - actionModifierResponseHeader:
      header:
        Server:
          - CatSync

actions:
  - route: ^/$
    type: string
    actionModifierStatus:
      status: 200
    string:
      actionModifierResponseHeader:
        header:
          Content-Type:
            - text/plain; charset=utf-8
      content: Hello

  - route: ^/secure$
    type: string
    actionModifierAuth:
      header:
        X-Token:
          - ^dev$
      fallback:
        type: next
    string:
      content: ok

  # notfound handler (last action)
  - route: ""
    type: string
    actionModifierStatus:
      status: 404
    string:
      content: "page not found\n"
```
