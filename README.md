# content-unlock

内容解锁：按等级门槛与前置内容判断哪些内容可以解锁，前置关系不许成环，只用 Go 标准库。

```
go run ./cmd/content-unlock --sample=unknown
go run ./cmd/content-unlock --sample=cycle
go run ./cmd/content-unlock --sample=level
go run ./cmd/content-unlock --sample=again
go run ./cmd/content-unlock --sample=dup
go run ./cmd/content-unlock --sample=scan
go test ./...
go vet ./...
```

## 口径（README 为准）

- **登记**：同一个 id 重复登记返回 false；前置里有没登记过的 id 返回 false；加了这条会形成环
  （含自环）也返回 false；三种情况都不改变图。
- **解锁条件**：等级不小于内容门槛，且全部前置都已解锁。
- **增量**：`Unlock` 反复推进到没有新解锁为止，只返回本次新增的 id（升序），已经解锁过的不再返回。
- **不变量**：已解锁集合单调不减；条件不满足的内容永不解锁；同一串输入重复执行结果相同。
- **代价**：`Unlock` 不许每次全扫内容表，`Scanned` 不随内容数乘查询次数增长；10 万条内容每秒 10 万次解锁。

## 输出契约（不改格式）

```
ok=..
count=..
second=..
scanned=..
```

## 目录

```
unlock.go            前置图、环检测、等级门槛与增量解锁
unlock_test.go       单元测试
cmd/content-unlock/  命令行入口
```
