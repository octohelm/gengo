---
name: gengo-guideline
description: 封装 octohelm/gengo 的自定义生成器扩展方式与项目接入约定；当任务涉及扩展生成器、注册 side import、组装执行入口或排查 +gengo 生成行为时使用。
---

# gengo-guideline

按 `github.com/octohelm/gengo` 约定接入代码生成或扩展自定义生成器。

## 接入项目

```go
package main

import (
    _ "github.com/octohelm/gengo/devpkg/deepcopygen" // side import 注册生成器
    "github.com/octohelm/gengo/pkg/gengo"
)

func main() {
    c, _ := gengo.NewExecutor(&gengo.GeneratorArgs{
        Entrypoint:         []string{"."},
        OutputFileBaseName: "zz_generated",
        Force:              true,
    })
    c.Execute(context.Background(), gengo.GetRegisteredGenerators()...)
}
```

**关键约定**：
- 生成器通过 `side import` + `init()` 自动注册
- `Entrypoint` 指向目标包路径，`OutputFileBaseName` 控制输出前缀
- 具体 API 以 `go doc github.com/octohelm/gengo/pkg/gengo` 为准

## 扩展自定义生成器

见 [references/extend-custom-generators.spec.md](references/extend-custom-generators.spec.md)——接口选择、触发规则、标签来源、测试分层。

## 项目接入更多细节

见 [references/use-gengo.spec.md](references/use-gengo.spec.md)——GeneratorArgs 完整配置、常见排查点。
