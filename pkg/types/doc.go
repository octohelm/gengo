/*
Package types 负责加载 Go 包，并暴露读取声明、注释指令、类型引用和函数返回值推导结果的辅助能力。

生成器可以通过这个包检查源码包，而不必手动遍历 go/packages 和 go/types 的底层数据结构。
*/
package types
