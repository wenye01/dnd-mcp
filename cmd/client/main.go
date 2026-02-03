// Package main 是 DND MCP Client 的主程序入口
package main

import (
	"os"

	"github.com/dnd-mcp/client/internal/cli"
)

func main() {
	cli.Execute()
	os.Exit(0)
}
