// Package agentskill 提供从 Go module 中安装 .agents/skills 的能力。
//
// 它读取 go.mod 中的 skill 引用（通过注释或 // +skill:{skill-name} 指令），计算安装计划，并将
// skill 目录从 module cache 软链到目标项目的 .agents/skills/ 目录。
//
// 常见入口：
//   - 用 ParseGoModSkillsFile 从 go.mod 解析 skill 引用列表。
//   - 用 PlanSkillInstall 根据 module cache 计算安装计划。
//   - 用 ApplyInstallPlan 执行安装。
package agentskill
