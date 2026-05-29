package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/skylunna/agent-runtime/adapters/llm"
	"github.com/skylunna/agent-runtime/ports"
)

func main() {
	apiKey := os.Getenv("DASHSCOPE_API_KEY")
	if apiKey == "" {
		log.Fatal("请设置环境变量 DASHSCOPE_API_KEY")
	}

	provider := llm.New(apiKey)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Chat(ctx, ports.ChatRequest{
		Model: "qwen-turbo", // 便宜快速
		Messages: []ports.Message{
			{Role: ports.RoleUser, Content: "用一句话介绍你自己"},
		},
	})
	if err != nil {
		log.Fatalf("调用失败: %v", err)
	}

	fmt.Println("=== 回复 ===")
	fmt.Println(resp.Content)
	fmt.Printf("=== 用量 === prompt=%d completion=%d\n",
		resp.PromptTokens, resp.CompletionTokens)
}
