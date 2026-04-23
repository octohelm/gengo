# 列出所有可用命令（无输入）
[group('meta')]
default:
    @just --list --list-submodules

# Go 工具链入口
[group: 'toolchain']
mod go 'tool/go'
