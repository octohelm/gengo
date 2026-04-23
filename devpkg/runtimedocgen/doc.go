// Package runtimedocgen 提供内置的 runtimedoc 生成器。
//
// 生成器名为 `runtimedoc`，通常通过类型注释上的 `+gengo:runtimedoc`
// 启用。
//
// 生成结果会把类型和字段注释转成运行时可查询的 RuntimeDoc 方法：
//
//   - struct 类型会生成 `RuntimeDoc(names ...string) ([]string, bool)`；
//   - 导出字段会按字段名生成查询分支；
//   - 嵌入字段会通过 helper 继续向下查询；
//   - 注释中的 `[[path]]` 会转换为 go:embed 字符串；
//   - 未导出类型、interface 类型和 alias 类型会被跳过。
//
// 该包通过 side import 触发 init 注册：
//
//	import _ "github.com/octohelm/gengo/devpkg/runtimedocgen"
package runtimedocgen
