# internal/testutil/tddcheck

本目录只保留本仓库的架构约束测试入口。

规则实现来自：

```go
github.com/lwmacct/260622-go-pkg-tddcheck/pkg/tddcheck
```

运行全部检查：

```bash
go test -count=1 ./internal/testutil/tddcheck
```

只运行架构检查入口：

```bash
go test -count=1 ./internal/testutil/tddcheck -run TestRules
```

这些测试会扫描仓库源码和目录结构，日常运行时必须带 `-count=1`，避免 Go test 缓存复用旧的架构检查结果。
