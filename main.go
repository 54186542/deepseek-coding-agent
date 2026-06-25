package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/54186542/deepseek-coding-agent/internal/llm"
	"github.com/54186542/deepseek-coding-agent/internal/platform"
	"github.com/54186542/deepseek-coding-agent/internal/react"
	"github.com/54186542/deepseek-coding-agent/internal/safety"
	"github.com/54186542/deepseek-coding-agent/internal/tool"
)

func init() {
	os.MkdirAll("spec", 0755)
	os.WriteFile(filepath.Join("spec", "v_001.md"), []byte("我想用wails为这个项目编写一个ui界面"), 0644)
}

func main() {
	var (
		apiKey  = flag.String("api-key", "", "DeepSeek API key (env: DEEPSEEK_API_KEY)")
		model   = flag.String("model", "deepseek-chat", "Model name")
		rootDir = flag.String("root", "", "Project root directory (default: current working directory)")
	)
	flag.Parse()

	// API key from flag or env
	key := *apiKey
	if key == "" {
		key = os.Getenv("DEEPSEEK_API_KEY")
	}
	if key == "" {
		fmt.Fprintln(os.Stderr, "错误: 未设置 API key。通过 --api-key 参数或 DEEPSEEK_API_KEY 环境变量设置。")
		os.Exit(1)
	}

	// Project root
	root := *rootDir
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取当前目录失败: %v\n", err)
			os.Exit(1)
		}
		root = wd
	}

	// Initialize layers
	adapter := platform.NewAdapter()
	sandbox := safety.NewSandbox(root)

	reg := tool.NewRegistry()
	tool.RegisterAll(reg, adapter, sandbox)

	client := llm.NewClient(key, *model)

	loop := react.NewLoop(client, reg, adapter)

	fmt.Printf("项目根目录: %s\n", adapter.ToSlash(root))
	if err := loop.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
