# gengo

用于 Go 代码生成、源码分析与生成结果组装的能力库。

- 整体架构 [ARCHITECTURE](docs/ARCHITECTURE.md)
- 项目约定 [CODING_GUIDELINE](docs/CODING_GUIDELINE.md)
- 领域上下文与术语表 [CONTEXT.md](./CONTEXT.md)（由 `skill:grill-with-docs` 维护）
- 架构决策记录 [docs/adr/](docs/adr/)（由 `skill:grill-with-docs` 维护）
- 环境与工具版本详见 [mise.toml](./mise.toml)
- 可用命令入口 [justfile](./justfile) 或 `just --list`

## 重要约束

- 环境变量涉及敏感信息，请勿自行探索；遇到问题且怀疑是环境变量时，先停下来询问
- 只修改与当前任务直接相关的文件，不因目录已存在而默认新增模块、入口或注册关系
- 不做与任务无关的功能扩展、顺手重构或批量风格清理
- 新增子模块、注册入口或改变执行面模式前，必须先确认
- 先做最小盘点，再开始改动
- 能完成最小验证时，必须先验证再交付；无法验证时说明原因与风险
- 文档、注释、错误或日志信息尽量使用中文（专有名词除外）
- 涉及领域概念时，先查阅 `CONTEXT.md` 中的术语表，使用其中定义的术语
