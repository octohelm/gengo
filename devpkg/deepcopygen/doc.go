// Package deepcopygen 提供内置的 deepcopy 生成器。
//
// 生成器名为 `deepcopy`，通常通过类型注释上的 `+gengo:deepcopy:*` 参数驱动：
//
//   - `+gengo:deepcopy`
//     为当前命名类型生成 DeepCopy 和 DeepCopyInto 方法。
//   - `+gengo:deepcopy:interfaces=<type>`
//     额外生成 DeepCopyObject 方法，并让返回值类型使用指定接口。
//
// 生成结果按类型底层结构展开：
//
//   - struct 会生成指针接收者的 DeepCopy 和 DeepCopyInto；
//   - map 会生成值接收者的 DeepCopy 和 DeepCopyInto；
//   - 本包内被字段引用的命名类型会递归补齐生成；
//   - interface 类型会被跳过。
//
// 该包通过 side import 触发 init 注册：
//
//	import _ "github.com/octohelm/gengo/devpkg/deepcopygen"
package deepcopygen
