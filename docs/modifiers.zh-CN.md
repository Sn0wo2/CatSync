# Modifiers 语义与约束

本文档描述 modifier 的运行时语义、组合方式与未来扩展约束。

## 组合顺序

当前顺序：

1. global modifiers
2. action modifiers
3. payload modifiers
4. type-specific modifiers（version placeholder）

顺序是稳定 API 的一部分；如果未来修改，需要在 release notes 中显式声明。

## 叠加 vs 覆盖

当前实现主要是“叠加”。例如 response headers 使用 fiber `Append`，同名 header 会追加。

如果需要“覆盖”，必须定义清晰规则：

- 是覆盖全局同名 header，还是覆盖某个 key 的值？
- 是否允许同时追加与覆盖？
- 当 global/action/payload 都配置时如何处理？

建议后续做法：

- 为每个 modifier 类型增加 resolution 语义（例如 `mode: append|replace`）
- 或在 builder 阶段合并并输出一个“最终态 modifier”，避免运行时多次叠加

## 副作用与可预测性

modifier 应尽量满足：

- 副作用局部化：只修改当前请求/响应
- 行为可预测：不依赖全局状态
- 失败可观测：返回 error，不 panic

auth jump 是目前唯一允许的控制流副作用，其它 modifiers 不应引入跳转能力。

## 约束

- modifier schema（config）与运行时 modifier（action）要保持一一对应
- 不在 config 里放运行时对象
- 所有 nil pointer 访问必须在执行器或 handler 内显式检查并返回 error
