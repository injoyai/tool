# Timer V2 项目笔记

## 脚本引擎
- 使用 `github.com/traefik/yaegi v0.16.1` 作为 Go 脚本解释器
- 内置函数包 `funcs` 位于 `script.go`,提供: Start, Ping, Notice, Dial, DialTCP, Print, Println
- 脚本必须是完整 Go 程序(`package main` + `func main`)
- 脚本可用标准库 + funcs 包 + lib 符号(conv/base/logs 等)
- yaegi CLI 已安装,重新 extract: `cd lib && yaegi extract <pkg>`

## 前端
- CodeMirror 5 编辑器,Go 语法高亮
- 静态文件在 `static/` 目录,通过 embed + `/static/*` 路由
- 默认脚本模板定义在 index.html 的 DEFAULT_SCRIPT

## 构建
- 入口:`cmd/server/main.go`(原 cmd/web 已改名),构建命令:`go build -o timer.exe ./cmd/server/`
- 支持命令行 flag:`-port`(默认8078)、`-db`(默认 ./data/database/timer.db)
- 运行脚本:`run.sh`(完整运行/dev预览)、`run.bat`(Windows双击)、`build.sh`(发布构建)
- Go 1.25 vet 兼容问题:测试用 `go test -vet=off`
- **路由参数注意**:gofiber 中 `c.Get(key)` 取请求头,取 wildcard 参数必须用 `c.Params("*")`
