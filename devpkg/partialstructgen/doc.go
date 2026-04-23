// Package partialstructgen 提供内置的 partialstruct 生成器实现。
//
// 生成器名为 `partialstruct`，通常通过类型注释上的
// `+gengo:partialstruct:*` 参数驱动。
//
// partialstruct 要求目标类型按 `type Target sourcepkg.Type` 形式声明，
// 生成器会基于原始类型字段生成一个可部分更新的结构体：
//
//   - `+gengo:partialstruct`
//     为当前类型启用 partialstruct 生成。
//   - `+gengo:partialstruct:omit=<field>`
//     从生成结构体和 DeepCopyIntoAs 中省略指定字段。可重复声明。
//   - `+gengo:partialstruct:replace=<field>:<type> [tag]`
//     替换指定字段的类型，并可选替换字段 tag。
//
// 生成结果包含：
//
//   - 与目标类型同名的部分结构体；
//   - `DeepCopyAs() *Origin`；
//   - `DeepCopyIntoAs(out *Origin)`。
//
// 该包通过 side import 触发 init 注册：
//
//	import _ "github.com/octohelm/gengo/devpkg/partialstructgen"
package partialstructgen
