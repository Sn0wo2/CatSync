# 文档

中文文档是主要参考。

## 索引

- [快速开始](#快速开始)
- [设计与架构](#设计与架构)
- [配置](#配置)
- [执行与控制流](#执行与控制流)

## 设计与架构

- [v2 设计总览](design-v2-modifiers-execution.zh-CN.md)（modifiers-first、执行流、auth fallback、notfound）
- [模块划分与依赖边界](architecture.zh-CN.md)（action/config/router/framework）

## 配置

### 快速开始

- [快速开始](quick-start.zh-CN.md)

### 指南

- [详细配置指南（action/modifier 逐项解释）](config-guide.zh-CN.md)
- [配置结构与字段含义（按 scope 与 modifier 分类）](config-schema.zh-CN.md)
- [默认配置示例说明与调试方式](default-config.zh-CN.md)
- [破坏性变更与迁移指引](migration-v2.zh-CN.md)

## 执行与控制流

- [请求匹配、执行器、jump 控制流、notfound](execution.zh-CN.md)
- [modifier 语义、顺序、可预期性与后续扩展点](modifiers.zh-CN.md)
