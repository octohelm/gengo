# 使用 gengo 规范

本文描述如何在项目中组装并执行 `gengo`，重点是执行入口、`GeneratorArgs` 语义和注册生成器的接线方式。

## 适用范围

适用于以下任务：

- 在项目里新增 `gengo` 执行入口
- 调整入口命令的 `GeneratorArgs`
- 增删 side import 的生成器注册
- 排查“为什么没有生成文件”或“为什么某些标签全局生效”

## 事实边界

截至当前仓库状态，`tool/` 下只有格式化与 skill 安装相关入口，没有现成的 `gengo` 生成命令可直接复用。

典型流程是：

1. side import 所需生成器，触发 `init()` 注册。
2. 调用 `gengo.NewContext(&gengo.GeneratorArgs{...})` 创建执行器。
3. 调用 `c.Execute(ctx, gengo.GetRegisteredGenerators()...)` 执行全部已注册生成器。

## 参考接线

若任务需要新增或审计项目接线，应以“按项目自建入口、沿用仓库运行时约定”为前提，而不是假设仓库已有固定命令。推荐入口形态如下：

```go
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"
)

func main() {
	flag.Parse()

	entrypoints := flag.Args()
	if len(entrypoints) == 0 {
		cwd, err := os.Getwd()
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
			return
		}

		entrypoints = append(entrypoints, cwd)
	}

	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint:         entrypoints,
		OutputFileBaseName: "zz_generated",
	})
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}
}
```

## `GeneratorArgs` 约定

### `Entrypoint`

- 接受 import path 或合法相对目录
- 项目型命令常直接把当前工作目录作为入口，例如 `cwd`

### `OutputFileBaseName`

- 控制生成文件前缀
- 生成文件名固定为 `<OutputFileBaseName>.<generator>.go`
- 项目里常见值是 `zz_generated`

### `Globals`

- 为所有包注入共享标签
- 适合像示例中那样，给某个生成器加全局开关，例如 `gengo:runtimedoc`
- 排查“为什么没有在源码注释里写标签却仍然生效”时，要优先检查这里

### `All`

- `false` 时只处理直接入口包
- `true` 时会遍历本地包，并结合 `gengo.sum` 做增量判断

### `Force`

- 为 `true` 时忽略缓存并强制重新生成

## 注册生成器的方式

执行入口本身通常不直接构造生成器，而是依赖 side import 后的全局注册表：

- `gengo.Register(...)` 负责注册
- `gengo.GetRegisteredGenerators()` 负责取回全部已注册生成器
- 如需只执行子集，可传名称筛选：`gengo.GetRegisteredGenerators("foo", "bar")`

若项目入口要混用 `gengo` 自带生成器和其它仓库的生成器，继续沿用 side import 即可。

## 入口实现建议

- 新入口优先保持为普通 Go `main` 包，不把项目局部约定硬编码进 `pkg/gengo` 运行时。
- 默认显式设置 `OutputFileBaseName`，避免不同项目沿用不同默认值却无从排查。
- 若入口允许命令行参数传入目标包，空参数时再决定是否回退到当前工作目录。
- 若需要只跑部分生成器，优先通过 `gengo.GetRegisteredGenerators("foo", "bar")` 过滤，而不是改注册表全局状态。

## 常见排查点

### 没有生成文件

优先检查：

- 入口目录是否传给了 `Entrypoint`
- 目标生成器是否通过 side import 注册
- 目标类型上是否存在 `+gengo:<name>` 或相应选项标签
- 是否被缓存或 `All` / `Force` 配置影响

### 某个标签全局生效

优先检查：

- `GeneratorArgs.Globals`
- 包级注释
- 声明级注释

## 验证建议

- 改入口命令：至少验证入口所在模块的生成命令和相关 `go test`
- 改 `GeneratorArgs` 语义理解：至少覆盖 `./pkg/gengo/...`
- 改 side import 组合：再检查受影响生成器对应包
- 若仓库内尚未存在目标入口，应在交付中明确“这是新增项目接线”，不要写成“修复现有仓库命令”。

如果本地无法直接运行项目生成命令，要明确缺口，不编造验证结果。
