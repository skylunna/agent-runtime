package ports

import "context"

// LLMProvider 抽象 [调一次大模型] 这件事
//
// 因为 DeepSeek、Qwen、OpenAI、以及未来的 Luner 都是 OpenAI 兼容接口
// 所以这个接口按 OpenAI 的 chat completions 形态来设计
// v1 实现直连 DeepSeek / Qwen，未来切回 Luner 只改 base_url，接口不动
type LLMProvider interface {
	// Chat 发起一次对话补全
	//
	// 关键：返回的 ChatResponse 会被原样包进LLMCallCompleted事件落盘
	// 重放时直接读这个罗盘的旧结果，绝不重新调用 LLM -
	// 这是 【确定性可重放】 的核心，也是 LLM 非确定性问题的解法
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}

// Role 是对话中一条消息的角色，沿用 OpenAI的约定
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool" // 工具执行结果回灌给模型时用
)

// Message 是对话里的一条消息
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 是一次补全请求。先保持最小，只放跑通单步所必需的字段
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse 是模型的返回。同样保持最小
type ChatResponse struct {
	Content string `json:"content"` // 模型生成的文本

	// 用量信息，罗盘后对成本统计、可观测都有用
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
