// Package testingutil 提供跨包生成器测试使用的临时模块与组合式断言辅助能力。
//
// 该包主要服务 devpkg 下的生成器测试。测试可以用 NewModule 构造隔离
// Go module，用 Module.Generate 执行生成器，再用 File、Contains 等
// cmp 风格谓词组合到 testing/v2 的 Be 断言中。
package testingutil
