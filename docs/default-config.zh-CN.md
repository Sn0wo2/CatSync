# 默认配置说明

默认配置由 `config/GetDefaultConfig()` 生成（`config/default.go`）。

设计目标：

- 展示主要功能点
- 启动即可用（尽量减少手动准备）
- 作为用户编写配置的参考模板

## 包含的示例路由

以下描述以默认配置为准：

- `/`
    - string action
    - 演示：status、response header、version placeholder

- `/public/hello.txt`
    - file action
    - 演示：file 读取、cache-control
    - 文件不存在或为空时会被自动创建并写入 `CatSync!\n`

- `/secure`
    - string action
    - 演示：auth header 校验 + fallback next

- `/secure/jump`
    - string action
    - 演示：auth 校验失败触发 fallback jumpTo

- `/secure/jump/fallback`
    - string action
    - 演示：作为 jump 的 fallback action 返回 401（该 action 本身不配置 auth）

## Not found

默认配置使用“最后一个 action 作为 notfound handler”的约定。

- status: 404
- msg: `page not found\n`
- Content-Type: text/plain; charset=utf-8

## 调试建议

### 验证 auth

- 通过：`curl -H "X-Token: dev" http://localhost:3000/secure`
- 失败（next）：`curl http://localhost:3000/secure`

### 验证 jump

- `curl http://localhost:3000/secure/jump`
    - 会跳到 `/secure/jump/fallback` 逻辑（由 index 驱动，不依赖 route 命中）
