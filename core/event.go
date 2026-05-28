package core

import "time"

// EventType 标识一个事件的种类
// 关键设计: 意图 (Requested) 和 结果 (Completed) 必须是不同的事件类型
// 成对出现: 这样崩溃恢复时才能区分 "做了一半" 和 "做完了"
type EventType string

const (
	TaskStarted       EventType = "task.started"
	LLMCallRequested  EventType = "llm.requested"  // 意图: 准备调 LLM
	LLMCallCompleted  EventType = "llm.completed"  // 结果: LLM 返回了
	ToolCallRequested EventType = "tool.requested" // 意图: 准备调用工具
	ToolCallCompleted EventType = "tool.completed" // 结果: 工具返回了
	TaskCompleted     EventType = "task.completed" // 结果: 任务完成了
	TaskFailed        EventType = "task.failed"    // 结果: 任务失败了
)

// Event 是不可变的。一旦产生并罗盘，用不修改
// 这是事件溯源的铁律：只append，不update，不delete
type Event struct {
	// 身份与顺序
	ID       string `json:"id"`                  // 唯一标识一个事件
	TaskId   string `json:"task_id"`             // 事件所属的任务
	Seq      int64  `json:"seq"`                 // 事件在任务中的顺序，从1开始递增
	ParentID string `json:"parent_id,omitempty"` // 父事件，用于重建执行树 (未来支持嵌套 / sub-agent)

	// 内容
	Type    EventType `json:"type"`    // 事件类型
	Payload []byte    `json:"payload"` // 事件内容，JSON序列化的任意结构

	// 元数据 (可观测性从这里派生)
	Timestamp time.Time `json:"timestamp"` // 事件发生的时间
}
