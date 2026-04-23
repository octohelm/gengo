/*
Package gengo 实现生成器运行时使用的核心 API。

它负责加载包、从注释中解析 `+gengo` 指令、协调生成器执行，并把格式化后的输出文件写回源码模块。

构建执行器应优先使用 NewExecutor。NewContext 作为兼容入口保留。
*/
package gengo
