# 配置指南（v2）

返回索引：[`docs/README.md`](README.md)

## 目录

- [顶层字段](#cg-top)
- [Action](#cg-action)
- [Action Types](#cg-action-types)
- [Modifiers（逐项解释）](#cg-modifiers)
    - [actionModifierResponseHeader](#m-response-header)
    - [actionModifierStatus](#m-status)
    - [actionVersionModifier](#m-version)
    - [actionModifierAuth](#m-auth)

本文档是面向用户的“详细配置手册”，按 action type 与 modifier 类型逐项说明。

配置类型定义在 `config/types.go`，本文以 YAML 为例。

## reader.String（通用字符串输入）

v2 beta 中，多数原本为 `string` 的字段统一改为 `reader.String`，目的是让配置可以“直接写字符串”，也可以“从文件/HTTP 拉取字符串”。

最常用写法（短写，默认字面量）：

```yaml
server:
  address: :3000
```

从文件读取（对象写法）：

```yaml
log:
  dir:
    type: path
    content: ./logs
```

从 HTTP(S) 读取（对象写法）：

```yaml
server:
  tls:
    cert:
      type: http
      content: https://example.com/tls/cert.pem
    key:
      type: http
      content: https://example.com/tls/key.pem
```

注意：短写字符串不会自动识别为 path/http；需要读取文件或 URL 时必须显式写 `type`。

<a id="cg-top"></a>

## 顶层字段

### log

```yaml
log:
  dir: ./logs
  level: debug
  fileFormat: 2006-01-02.log
```

- `dir`: 日志目录
- `level`: debug/info/warn/error
- `fileFormat`: 日志文件名格式

### server

```yaml
server:
  address: :3000
  tls:
    cert: ""
    key: ""
```

- `address`: 监听地址（reader.String）
- `tls.cert`/`tls.key`: 证书路径（reader.String），同时提供时启用 TLS

备注：`server.header` 在 v2 已移除。
如需设置 `Server` 响应头，请使用 `actionModifierResponseHeader`。

### modifiers（全局）

```yaml
modifiers:
  - actionModifierResponseHeader:
      header:
        Server:
          - CatSync
```

这是全局 modifiers 列表，对所有命中的 action 生效。

### actions

```yaml
actions:
  - route: ^/$
    type: string
    string:
      content: Hello
```

actions 按顺序扫描（index 是可观察行为，`jumpTo` 也是基于 index）。

<a id="cg-action"></a>

## Action

每个 action 必需字段：

- `route`: regexp
- `route`: reader.String（通常短写即可）
- `type`: `string` 或 `file`
- 对应的 payload 块：
    - `type: string` -> 必须有 `string:`
    - `type: file` -> 必须有 `file:`

action 还可以直接配置 modifiers（action scope）：

```yaml
actions:
  - route: ^/example$
    type: string
    actionModifierStatus:
      status: 200
    actionModifierResponseHeader:
      header:
        Content-Type:
          - text/plain
    string:
      content: ok
```

### route 为空的语义（重要）

`route` 允许为空字符串。

- `route: ""` 时，该 action 不参与正常的路径匹配扫描。
- 该 action 只能通过 `jumpTo` 进入执行（jump-only 节点）。

注意事项：

- 这类 action 不应该被期望能通过 `curl /some/path` 直接访问。
- 如果你把一个 action 的 route 置空但又希望它能作为 notfound/fallback：直接把它放到 actions 的最后一个。

notfound 约定（v2）：

- router 永远会无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）
- 因此建议最后一个 action：
    - 显式配置 `actionModifierStatus: 404`
    - 谨慎处理 `file`/鉴权，避免泄露敏感内容

<a id="cg-action-types"></a>

## Action Types

### string action

```yaml
actions:
  - route: ^/$
    type: string
    string:
      content: "Hello, CatSync!\n"
```

- `content`: 返回的字符串
- `content`: reader.String

string payload 也可以携带 modifiers（payload scope）：

```yaml
string:
  actionModifierResponseHeader:
    header:
      Content-Type:
        - text/plain; charset=utf-8
  content: ok
```

### file action

```yaml
actions:
  - route: ^/public/hello\\.txt$
    type: file
    file:
      path: ./data/hello.txt
      dontSetContentType: false
```

- `path`: reader.String 文件路径（必须在 `./data` 目录下）
- `dontSetContentType`: 为 true 时不自动设置 Content-Type

运行时行为（当前实现）：

- 文件不存在：创建文件并写入 `CatSync!\n`
- 文件存在但为空：写入 `CatSync!\n`

<a id="cg-modifiers"></a>

## Modifiers（逐项解释）

Modifiers 在 global/action/payload 三个 scope 的写法一致。

<a id="m-response-header"></a>

### actionModifierResponseHeader

追加响应头：

```yaml
actionModifierResponseHeader:
  header:
    Cache-Control:
      - public, max-age=60
```

- 使用 fiber 的 `Append`，因此同名 header 会追加多个值。

<a id="m-status"></a>

### actionModifierStatus

设置响应状态码：

```yaml
actionModifierStatus:
  status: 204
```

有效范围：100-599。

如果不配置 `actionModifierStatus`：

- string/file handler 通常会返回框架默认的 200（除非其它中间件/handler 修改了状态码）。
- 对 notfound 等兜底响应，是否返回 404 取决于最后一个 action 是否配置了 `actionModifierStatus: 404`。

<a id="m-version"></a>

### actionVersionModifier

版本占位符替换：

```yaml
actionVersionModifier:
  placeholder: ${VERSION}
```

当前主要用于响应头值（以及可选的 header key）替换。

<a id="m-auth"></a>

### actionModifierAuth

请求鉴权：

```yaml
actionModifierAuth:
  header:
    X-Token:
      - ^dev$
  query:
    user:
      ^[a-z0-9_]+$
  fallback:
    type: next
```

- `header`: map[string][]reader.String，每个值是 regexp
- `query`: map[string]reader.String，值是 regexp
- `ipAllowlist`: 可选，IP/CIDR 白名单（匹配任意一个则通过）
- `ipAllowlistFile`: reader.String，可选，从 file/http/string 读取 IP/CIDR 白名单（每行一个；支持 # 注释）
- `fallback`:
    - `type: next`: 校验失败时 `ctx.Next()`
    - `type: jump`: 校验失败时跳到 `jumpTo` 指定的 action index

要求（v2）：

- 只要配置了 `actionModifierAuth`，就必须显式配置 `fallback`（启动时会校验；缺失会直接报错退出）。

jump 示例（建议 jump 到一个明确的 fallback action，避免循环）：

```yaml
actionModifierAuth:
  header:
    X-Token:
      - ^dev$
  fallback:
    type: jump
    jumpTo: 4
```
