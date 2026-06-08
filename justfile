# 列出所有可用命令（无输入）
[group('meta')]
default:
    @just --list --list-submodules

mod go 'tool/go'
