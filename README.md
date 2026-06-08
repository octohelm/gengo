# gengo

用于 Go 代码生成、源码分析与生成结果组装的能力库。

它提供生成器执行骨架与一组围绕命名、类型加载、格式化和文件组织的通用支撑，目标是把生成流程里的公共问题收敛为可复用库，而不是在各个生成器里重复实现。

## 文档导航

| 文档 | 用途 | 受众 |
|------|------|------|
| [README.md](./README.md)（本文件） | 项目介绍、快速开始、职责边界 | 开发者 |
| [AGENTS.md](./AGENTS.md) | Agent 行为约束与控制面边界 | AI Agent |
| [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) | 系统拓扑、数据流、部署视图 | 开发者 |
| [docs/CODING_GUIDELINE.md](./docs/CODING_GUIDELINE.md) | 编码约定与风格规范 | 开发者 |
| [CONTEXT.md](./CONTEXT.md) | 领域术语表（由 `skill:grill-with-docs` 维护） | Agent / 开发者 |
| [docs/adr/](./docs/adr/) | 架构决策记录（由 `skill:grill-with-docs` 维护） | 开发者 |

## 快速开始

```
just go test  # 运行测试
just go lint  # 代码检查（需要 golangci-lint）
```

- 查看仓库统一入口：`just`
- 先阅读核心能力：[`pkg/gengo`](./pkg/gengo)

## 职责与边界

- `pkg/gengo` 提供生成器注册、执行与生成文件组装相关能力。
- `pkg/*` 提供命名、格式化、类型加载、词形处理与 sumfile 等通用支撑。
- `devpkg/*` 放置开发期生成器或实验性扩展，不作为仓库根执行入口。
- `tool/internal/cmd/fmt` 提供仓库内部使用的 Go 工具入口。

## 入口索引

- [pkg](./pkg) 核心库与通用能力的主入口。
- [devpkg](./devpkg) 开发期生成器与实验性扩展样例。
- [tool](./tool) 工具链与仓库内部命令实现。
- [justfile](./justfile) 仓库级统一执行入口。
