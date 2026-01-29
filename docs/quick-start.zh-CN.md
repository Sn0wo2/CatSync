# 快速开始

返回索引：[`docs/README.md`](README.md)

## 目录

- [获取与运行](#qs-run)
- [配置文件位置](#qs-config-path)
- [验证默认配置](#qs-verify)
- [下一步](#qs-next)

本文档面向第一次启动 CatSync v2 的用户，目标是“能跑起来 + 能看懂默认行为”。

<a id="qs-run"></a>

## 获取与运行

### 二进制

从 release 下载后直接运行：

```bash
./CatSync
```

默认会监听 `:3000`。

### Docker

```bash
docker run -d -p 3000:3000 -v ./data:/app/data --name catsync ghcr.io/sn0wo2/catsync:beta-latest
```

注意：容器内配置默认位于 `/app/data/config.yml`（项目会在找不到时自动创建）。

<a id="qs-config-path"></a>

## 配置文件位置

加载顺序（简化描述）：

1. `CONFIG_PATH`
2. debug 模式下的 `DEBUG_CONFIG_PATH`
3. 默认搜索 `./data/config.yml|yml|json`

当指定的路径不存在时，程序会创建一份默认配置到该路径。

<a id="qs-verify"></a>

## 验证默认配置

默认配置包含以下路由示例：

- `/`
- `/public/hello.txt`
- `/secure`
- `/secure/jump`

### 访问首页

```bash
curl http://localhost:3000/
```

### 访问 file action

```bash
curl http://localhost:3000/public/hello.txt
```

如果 `./data/hello.txt` 不存在或为空，服务会创建并写入 `CatSync!`。

### 验证 auth

通过：

```bash
curl -H "X-Token: dev" http://localhost:3000/secure
```

失败（fallback=next，继续扫描后续 action，直到最后一个 notfound handler）：

```bash
curl http://localhost:3000/secure
```

### 验证 auth jump

```bash
curl http://localhost:3000/secure/jump
```

该路由在 auth 失败时会触发 `jumpTo`，跳到一个兜底 action。

<a id="qs-next"></a>

## 下一步

- 配置结构与字段解释：[`docs/config-guide.zh-CN.md`](config-guide.zh-CN.md)
- 默认配置逐项说明：[`docs/default-config.zh-CN.md`](default-config.zh-CN.md)
- modifiers 语义与组合：[`docs/modifiers.zh-CN.md`](modifiers.zh-CN.md)
