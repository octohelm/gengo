# gengo

Go 代码生成与相关源码处理能力库。

## 职责与边界

- `pkg/gengo` 提供生成器注册、执行与生成文件组装相关能力。
- `pkg/*` 提供命名、格式化、类型加载、词形处理与 sumfile 等通用支撑。
- `devpkg/*` 放置开发期生成器或实验性扩展，不作为仓库根执行入口。
- `tool/internal/cmd/fmt` 提供仓库内部使用的 Go 工具入口。

## 目录索引

- [pkg](./pkg) 核心库与通用能力
- [devpkg](./devpkg) 开发期生成器与扩展
- [tool](./tool) 工具链相关实现
- [justfile](./justfile) 仓库统一执行入口

## 最小入口

- 查看可用入口：`just`
- 运行全仓测试：`just go test ./...`
- 运行静态检查：`just go vet ./...`
