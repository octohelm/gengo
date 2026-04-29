---
name: gengo-guideline
description: 封装 octohelm/gengo 的自定义生成器扩展方式与项目接入约定；当任务涉及扩展生成器、注册 side import、组装执行入口或排查 +gengo 生成行为时使用。
---

# gengo-guideline

用于在这个仓库里稳定使用 `gengo` 运行时、生成器扩展约定与项目接线方式的 Tool Wrapper skill。

## 使用范围

- 适用于扩展或修改自定义生成器、补 side import 注册、组装执行入口、排查 `+gengo:*` 标签生效与生成结果异常。
- 不适用于普通业务代码修改、与 `gengo` 无关的工具链整理，或脱离本仓库事实去编造项目接线方式。

## 必要输入

- 当前任务落点是生成器实现、运行时接口，还是项目接线。
- 受影响的生成器名、标签名或目标入口包。
- 若任务涉及验证，需先确认能运行的最小 `go test` 范围。

## 关键约定

- 本 skill 只处理当前仓库内的 `gengo` 运行时、生成器、执行入口和相关生成产物。
- 扩展生成器时优先沿用 side import + `init()` 注册模式，除非任务明确要求改变入口组织。
- 不把不存在的仓库入口、命令或生成流程当作既成事实；需要先以代码搜索结果为准。
- 生成器测试优先复用仓库现有 `pkg/gengo/testingutil` 与包内测试样式。
- 仓库级协作规则仍以根目录 `AGENTS.md` 为准。

## 资源导航

- 要扩展、修改或测试自定义生成器，读 [`references/extend-custom-generators.spec.md`](references/extend-custom-generators.spec.md)。
- 要在项目里接入、执行或排查 `gengo` 入口，读 [`references/use-gengo.spec.md`](references/use-gengo.spec.md)。

## 执行与验证

- 先按任务类型选择 reference，再回到代码定位真实入口、生成器实现和现有测试。
- 若只改生成器实现，优先验证对应包测试；若改运行时标签语义或注册行为，至少补 `pkg/gengo` 相关测试。
- 若仓库内不存在声称的入口或脚本，应停止把它当作前提，并在交付中明确指出。

## 完成标准

- 已选择正确的 reference 作为当前任务入口。
- `SKILL.md` 只保留使用边界、关键约定和资源导航，不重复展开实现细节。
- 具体扩展规则、接线方式和测试写法均由对应 reference 承载。
- 交付时能说明实际验证范围，以及未验证项与风险。
