package ports

import "context"

type ChatRequest struct {
	// Add request fields here
}

type ChatResponse struct {
	// Add response fields here
}

type LLMProvider interface {
	// Chat 发起一次对话补全。 OpenAI 兼容，所以 v1 直接指向 Luner
	// 返回的内容会被包进去 LLMCallCompleted 事件罗盘，供重放使用
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
