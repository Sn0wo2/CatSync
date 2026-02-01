# 配置结构（v2）

本文档按“作用域 + modifier 类型”解释配置 schema，目标是让用户在不读 Go 代码的前提下能写出正确配置。

配置类型定义在 `config/types.go`。

## reader.String（通用字符串输入）

多数原本是 `string` 的字段在 v2 beta 里统一升级为 `reader.String`，以支持：

- 直接写字符串（默认）：`foo`
- 从文件读取：
  - `type: path`
  - `content: ./path/to/file`
- 从 HTTP(S) 读取：
  - `type: http`
  - `content: https://example.com/data`

短写形式（YAML/JSON 直接写一个字符串）默认按字面量处理，不会自动当作文件路径或 URL。
如需从文件/HTTP 读取，必须使用对象形式显式指定 `type`。

## 顶层结构

- `log`
    - `dir`: reader.String，日志目录
    - `level`: reader.String，日志级别
    - `fileFormat`: reader.String，日志文件名格式

- `server`
    - `address`: reader.String，监听地址，例如 `:3000`
    - `tls.cert`/`tls.key`: reader.String，TLS 证书路径（可选）
    - `tls.redirectHttp`: 非 challenge 请求是否 301 重定向到 https（可选；默认 true，设为 false 关闭；仅 ACME http-01 有意义）
    - `acme`（可选）：自动签发证书（ACME / Let's Encrypt）
        - `enable`: 是否启用
        - `hosts`: 允许签发的域名列表（启用时必填）
        - `email`: 可选，用于 Let's Encrypt 通知
        - `cacheDir`: 证书缓存目录（默认 `./data/acme`）
        - `directoryURL`: reader.String，可选，ACME Directory URL（可用于 staging）
        - `http01`（可选）：HTTP-01 验证配置（与 dns01 互斥；两者都不配则默认使用 http01）
            - `httpAddress`: reader.String，HTTP-01 challenge 监听地址（默认 `:80`）
        - `dns01`（可选）：DNS-01 验证配置（与 http01 互斥）
            - `provider`: reader.String，`exec`（默认）或 `cloudflare`/`dnspod`/`alidns`/`route53`（需要编译 tag 才会内置）
            - `presentCmd`: 添加 TXT 记录的命令（argv 数组）
            - `cleanupCmd`: 删除 TXT 记录的命令（argv 数组）
            - `propagationTimeoutSeconds`: 传播超时（秒）
            - `pollingIntervalSeconds`: 轮询间隔（秒）

说明：

- 发布版/默认全功能构建：使用 `-tags catsync_all`（包含 ACME http-01、dns-01 以及常见 DNS provider）
- 需要最小二进制：不加 tag 或仅开启需要的 feature/provider
- `provider=exec`：使用外部命令更新 TXT 记录（通用、依赖最少，二进制也最小）
- 其他 provider：使用 lego 内置 provider 直接调用云厂商 API，通常通过环境变量提供凭据；为了控制编译体积，这些 provider 需要通过 build tags 显式启用，例如 `-tags dns_cloudflare`

- `modifiers`（可选）
    - 全局 modifiers 列表（作用于所有命中的 action）

- `actions`
    - action 列表（按顺序扫描）

## Action

每个 action 结构：

- `route`: reader.String，regexp，用于匹配 `ctx.Path()`
- `type`: `string` 或 `file`
- modifiers（action scope，直接写在 action 下，不再有 `globalmodifier:` 包裹）
- payload：
    - `string.content`: reader.String payload
    - `file.path`: reader.String 文件路径

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

- `header`: map[string][]reader.String，每个 header 允许多个 regexp
- `query`: map[string]reader.String
- `ipAllowlist`: 可选，IP/CIDR 白名单（匹配任意一个则通过）
- `ipAllowlistFile`: reader.String，可选，从 file/http/string 读取 IP/CIDR 白名单（每行一个；支持 # 注释）
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

- placeholder: reader.String
- placeholder 替换发生在响应头阶段（当前实现针对 response headers）

## Not found（约定）

当所有 actions 都未命中时：router 会无条件执行最后一个 action 作为 notfound handler（跳过 route 匹配）。

建议：

- 把最后一个 action 的 `route` 设为空字符串（jump-only），避免它在正常扫描中被误匹配
- 在最后一个 action 上显式配置 `actionModifierStatus: 404`

## Build tags（按需裁剪功能）

为了控制二进制大小，一些功能可以通过 build tags 按需编译：

- 全功能构建：`-tags catsync_all`
- dotenv（.env）：`-tags feature_dotenv`
- config loader（yaml）：`-tags feature_config_yaml`
- config loader（json）：`-tags feature_config_json`
- ACME（http-01）：`-tags feature_acme_http01`
- ACME（dns-01）：`-tags feature_acme_dns01`
