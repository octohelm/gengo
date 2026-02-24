# 列出所有可用命令（无输入）
[group('meta')]
default:
    @just --list

# 运行全量单测（输入：args）
[group('test')]
test path *args:
    @go test -failfast --count=1 {{ args }} {{ path }}

# 运行全量单测（输入：args）
[group('test')]
vet path *args:
    @go vet {{ args }} {{ path }}

# 格式化当前仓库 Go 文件（无输入）
[group('fmt')]
fmt path:
    go tool fmt {{ path }}

# 整理依赖（无输入）
[group('env')]
dep:
    go mod tidy

# 更新直接依赖（无输入）
[group('env')]
update:
    go get -u ./...
