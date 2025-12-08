## CatSync TODO(V2)

`V2`将会重构现有`Action`架构, 会有`Break Changes`, 扔掉`V1`的思维镣铐

---

 - [ ] 解耦`Action`, 使得一个`Action Types`负责大部分事务甚至可以进行两个一起的操作, 以及可以选择性增加设置选项(例如reload)所有而不是区分成不同Action