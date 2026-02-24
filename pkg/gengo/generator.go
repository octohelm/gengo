package gengo

import (
	"errors"
	"go/types"
)

var (
	// ErrSkip 表示执行器应静默跳过当前声明。
	ErrSkip = errors.New("skip")
	// ErrIgnore 表示停止当前生成路径，但不把它视为失败。
	ErrIgnore = errors.New("ignore")
)

// GeneratorArgs 描述一次生成执行的包加载和输出命名配置。
type GeneratorArgs struct {
	// Globals 保存所有包共享的标签。
	Globals map[string][]string
	// Entrypoint 是 import path 或合法的相对目录路径。
	Entrypoint []string
	// OutputFileBaseName 是生成文件名的前缀。
	OutputFileBaseName string
	// All 为 true 时会处理所有依赖包。
	All bool
	// Force 为 true 时会忽略缓存并强制重新生成。
	Force bool
}

// Generator 为执行器选中的命名类型生成代码。
type Generator interface {
	// Name 返回启用当前生成器时使用的指令名。
	Name() string
	// GenerateType 为命名类型生成代码。
	GenerateType(Context, *types.Named) error
}

// AliasGenerator 为类型别名生成代码。
type AliasGenerator interface {
	// Name 返回启用当前生成器时使用的指令名。
	Name() string

	// GenerateAliasType 为别名类型生成代码。
	GenerateAliasType(Context, *types.Alias) error
}

// GeneratorNewer 会为特定 Context 创建一个生成器实例。
type GeneratorNewer interface {
	// New 返回绑定到 c 的生成器实例。
	New(c Context) Generator
}

// GeneratorCreator 允许在执行前自定义生成器初始化过程。
type GeneratorCreator interface {
	Init(Context, Generator, ...GeneratorPostInit) (Generator, error)
}

// GeneratorPostInit 会在生成器初始化完成后执行。
type GeneratorPostInit = func(g Generator, sw SnippetWriter) error
