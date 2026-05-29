package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/skylunna/agent-runtime/ports"
)

// 默认指向Qwen的 OpenAI 兼容端点
// 未来切 DeepSeek / OpenAI / Luner 只需要切换这个 baseURL
const defaultQwenBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"

// OpenAICompatible 是一个面向所有 OpenAI 兼容端点的 LLMProvider 实现
// 名字刻意不叫Qwen —— 它对 DeepSeek / OpenAI / Luner 同样适用
type OpenAICompatible struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// Option 用函数式选项模式做配置，方便后续扩展 (超时、自定义 client 等)
// 而不破坏构造函数签名
type Option func(*OpenAICompatible)

func WithBaseURL(url string) Option {
	return func(o *OpenAICompatible) { o.baseURL = url }
}

func WithHTTPClient(c *http.Client) Option {
	return func(o *OpenAICompatible) { o.client = c }
}

// New创建一个 OpenAI兼容的LLMProvider
// 默认指向Qwen 通过 WithBaseURL 可指向任意 OpenAI 兼容端点
func New(apiKey string, opts ...Option) *OpenAICompatible {
	o := &OpenAICompatible{
		baseURL: defaultQwenBaseURL,
		apiKey:  apiKey,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// 下面是OpenAI 兼容协议的请求 / 响应内部结构
// 故意不复用 ports.ChatRequest/ChatResponse，因为
// 1）ports是我们自己的 (领域类型) 要保持稳定，不该被外部协议污染
// 2）这里这些是 【适配层】 的私有类型，专门贴合OpenAI协议格式
// 这层翻译就是 适配器 的本职工作
type openAIRequest struct {
	Model   string          `json:"model"`
	Message []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message openAIMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// Chat 实现 prots.LLMProvider 接口
func (o *OpenAICompatible) Chat(ctx context.Context, req ports.ChatRequest) (ports.ChatResponse, error) {
	// 1. 把领域类型翻译成 OpenAI 协议类型
	msgs := make([]openAIMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = openAIMessage{Role: string(m.Role), Content: m.Content}
	}
	body, err := json.Marshal(openAIRequest{Model: req.Model, Message: msgs})
	if err != nil {
		return ports.ChatResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	// 2. 构造 HTTP 请求，带 ctx (支持超时 / 取消 是 Go 的好习惯)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return ports.ChatResponse{}, fmt.Errorf("build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)

	// 3. 发送
	httpResp, err := o.client.Do(httpReq)
	if err != nil {
		return ports.ChatResponse{}, fmt.Errorf("do request: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return ports.ChatResponse{}, fmt.Errorf("read response: %w", err)
	}

	// 4. 非 2xx 把原始 body 带回去 —— 调试时这一句能省很多事
	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return ports.ChatResponse{}, fmt.Errorf("llm api error: status=%d body=%s",
			httpResp.StatusCode, string(respBody))
	}

	// 5. 解析并翻译回领域类型
	var apiResp OpenAIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return ports.ChatResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(apiResp.Choices) == 0 {
		return ports.ChatResponse{}, fmt.Errorf("llm returned no choices")
	}

	return ports.ChatResponse{
		Content:          apiResp.Choices[0].Message.Content,
		PromptTokens:     apiResp.Usage.PromptTokens,
		CompletionTokens: apiResp.Usage.CompletionTokens,
	}, nil
}
