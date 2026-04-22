# gengo

[![GoDoc Widget](https://godoc.org/github.com/octohelm/gengo?status.svg)](https://pkg.go.dev/github.com/octohelm/gengo)

一个基于 Go 注释指令的模板代码生成框架，用 `+gengo:{name}[:args]` 风格标签驱动生成器发现、类型遍历和代码输出。

仓库提供核心运行时、内置生成器和配套测试样例，既可直接复用，也可作为自定义生成器实现参考。

## 内容总览

- [`pkg/gengo`](pkg/gengo)：核心生成框架，包括 generator 注册、上下文、错误约定和执行流程。
- [`devpkg`](devpkg)：仓库内置的生成器实现，通常通过 side import 方式触发 `init` 注册。
- [`pkg/types`](pkg/types)：注释标签解析等基础能力，包含 `+gengo:` 指令抽取的测试用例。
- [`pkg/format`](pkg/format)：生成结果格式化与 import 排序相关能力。
- [`internal/cmd/fmt`](internal/cmd/fmt)：仓库自用的格式化命令入口。
- [`testdata`](testdata)：用于验证注释指令和生成行为的样例输入。

## 内置生成器 side import

```go
import (
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/defaultergen"
	_ "github.com/octohelm/gengo/devpkg/partialstruct"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
)
```

`devpkg/*` 下每个包都提供了 `doc.go`，用于说明 side import 的使用方式和包职责。

## 自定义生成器示例

```go
package customgen

import (
	"go/ast"
	"go/types"

	"github.com/octohelm/gengo/pkg/gengo"
)

func init() {
	gengo.Register(&customGen{})
}

type customGen struct{}

func (*customGen) Name() string {
	return "custom"
}

func (g *customGen) GenerateType(c gengo.Context, named *types.Named) error {
	if !ast.IsExported(named.Obj().Name()) {
		return gengo.ErrSkip
	}

	if whenSomeThing() {
		return gengo.ErrIgnore
	}

	return nil
}
```

## 相关文档

- [`AGENTS.md`](AGENTS.md)：仓库级协同约束、停止条件与变更边界。
- [`go.mod`](go.mod)：模块依赖与当前仓库使用的 Go 版本。
