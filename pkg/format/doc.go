/*
Package format 负责格式化生成后的 Go 源码，并整理 gengo 输出中的 import 分组。

它会从 go.mod 注释中提取 import 分组规则，并在重写源码或遍历项目目录时应用这些规则。
*/
package format
