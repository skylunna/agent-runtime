package ports

import "context"

type ToolCall struct {
	// Name 是要调用的工具标识
	Name string
	// Params 为工具调用的参数
	Params map[string]interface{}
}

type ToolResult struct {
	// Data 为工具执行返回的数据
	Data interface{}
}

// ToolExecutor 执行工具调用。
type ToolExecutor interface {
	// Execute 执行一个工具调用。
	// v1 实现：本地 Go 函数（local.go）
	// v2 实现：容器/WASM 沙箱（sandbox.go）—— 接口不变
	Execute(ctx context.Context, call ToolCall) (ToolResult, error)
}
