---
name: gengo-guideline
description: 封装 octohelm/gengo 的自定义生成器扩展方式与项目接入约定；当任务涉及扩展生成器、注册 side import、组装执行入口或排查 +gengo 生成行为时使用。
---

# gengo-guideline

用于在这个仓库里处理 `gengo` 代码生成相关任务的 Tool Wrapper skill。

## 使用入口

- 要扩展或修改自定义生成器，读 [
  `references/extend-custom-generators.spec.md`](references/extend-custom-generators.spec.md)。
- 要在项目里接入、执行或排查 `gengo`，读 [`references/use-gengo.spec.md`](references/use-gengo.spec.md)。

## 边界

- 本 skill 只处理当前仓库内的 `gengo` 运行时、生成器、执行入口和相关生成产物。
- 生成器扩展细节放在 `references/extend-custom-generators.spec.md`。
- 项目接入与执行入口细节放在 `references/use-gengo.spec.md`。
- 仓库级协作规则仍以根目录 `AGENTS.md` 为准。

## 完成标准

- 已选择正确的 reference 作为当前任务入口。
- `SKILL.md` 只承担导航职责，不重复展开实现细节。
- 具体规则、接线方式和验证建议均由对应 reference 承载。
