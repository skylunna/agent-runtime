package ports

import "context"

// ToolExecutor 抽象 [执行一次工具调用] 这件事
//
// v1 实现：本地 Go 函数 (adapters/tool/local.go) —— 直接在进程里跑
// v2 实现: 容器 / WASM 沙箱 (adapters/tool/sandbox.go) —— 隔离执行不可信代码
// 切实实现时上层不动，这是把 [沙箱] 做成可插拔的关键

// 牢记非目标：我们绝不自己实现沙箱 v2 用现成的容器 / WASM 运行时做隔离
// 这个接口只负责 [编排] —— 调用、传参、收结果，不碰隔离机制本身

type ToolExecutor interface {
	// Execute 执行一个工具调用
	//
	// 和 LLM 一样：结果被包进 ToolCallCompleted 事件落盘
	// 但工具和 LLLM 有个关键区别 —— 工具往往有副作用 （写文件、发请求、该数据）
	// 所以重放时的处理更微妙：不能盲目重做
	Execute(ctx context.Context, call ToolCall) (ToolResult, error)
}

// ToolCall 是一次工具调用的请求
type ToolCall struct {
	Name string `json:"name"` // 要调用哪个工具
	Args []byte `json:"args"` // 调用参数，JSON 序列化的字节
}

// ToolResult 是工具执行的结果
type ToolResult struct {
	Content string `json:"content"`  // 工具返回的内容 (回灌给LLM)
	IsError bool   `json:"is_error"` // 工具执行是否出错
}
