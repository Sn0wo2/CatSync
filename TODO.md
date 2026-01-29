## CatSync TODO(V2)

`V2`将会重构现有`Action`架构, 会有`Break Changes`, 扔掉`V1`的思维镣铐  
我们将不在考虑维护`V1`版本, 请尽快根据配置文件的更新逐步过渡到`V2`

---

- [X] 解耦`Action`, 使得一个`Action Type`负责大部分事务甚至可以多个同时执行的操作, 以及可以选择性增加设置选项(
  例如`StringHandler`是可以同时进行热更新)所有而不是区分成不同`Action Type`
- [ ] 使用可重复利用的封装基础库, 不重复造轮子
- [X] 不再考虑`Yaml loader`的`Save`注释保留和合并新增字段, 没有填写的数值将直接使用默认数值(Default config or empty
  value) ~~维护这个太麻烦了, 或者后续考虑写在基础库~~
- [X] 使用正则表达式进行所有`Action`的`Auth`架构匹配
- [ ] 添加`WebHook API`, 尽可能保证自由度较高的配置, 后续用于自定义功能实现(例如: 统计)
- [ ] 添加热重启加载配置文件
- [ ] 增加`Action`的`File`缓存, 而不是每次请求都进行IO访问到磁盘读取数据(缓存策略可选内部LRU/Redis)
- [ ] 考虑添加热更新支持
- [X] 删除`api_health`的支持, 作为向后兼容请使用action作为替代
- [X] `Action` 可配置的状态码
- [ ] `Action` 可配置的限流
- [ ] 考虑将`Action`的`Auth`鉴权改为用户(组)体系架构, 简化权限管理
- [X] `Fallback action`支持自定义`action`而不是`404 not found`(最后一个 action 作为 notfound handler)
- [X] 把 _Me0wo_ 变成可爱猫娘
- [ ] 添加 `Action Type`: page
- [ ] 添加 `Action Type` proxy: 反向代理Action(Rewrite: header、body, 健康检查)
- [ ] 可选的`CORS`配置