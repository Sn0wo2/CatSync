## CatSync TODO(V2)

`V2`将会重构现有`Action`架构, 会有`Break Changes`, 扔掉`V1`的思维镣铐

---

- [X] 解耦`Action`, 使得一个`Action Types`负责大部分事务甚至可以进行两个一起的操作, 以及可以选择性增加设置选项(
  例如reload)所有而不是区分成不同`Action`
- [ ] 和其他项目一样, 引入自己写的封装基础库, 不重复造轮子
- [ ] 不再考虑`Yaml loader`的`Save`注释保留和合并新增字段, 没有填写的数值将直接使用默认数值(Default config or empty
  value)
- [X] 使用正则表达式进行所有`Action`的`Auth`架构匹配
- [ ] 添加`WebHook API`, 尽可能保证自由度较高的配置, 后续用于统计
- [ ] 添加热重启加载配置文件
- [ ] 增加`Action`的`File`缓存, 而不是每次请求都进行IO访问到磁盘读取数据(缓存策略可选内部LRU/Redis)
- [ ] 考虑添加热更新支持
- [ ] 删除`api_health`的支持, 作为向后兼容请使用action作为替代
- [ ] `Action` 可配置的状态码
- [ ] `Action` 可配置的限流
