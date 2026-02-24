/*
Package namer 负责把类型引用转换成源码层面的标识符，并保持 import 一致性。

它主要服务于 gengo 的 snippet 渲染，用来为导入包选择本地名称，并按生成目标包的上下文输出类型名。
*/
package namer
