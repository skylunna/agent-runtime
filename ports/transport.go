package ports

import (
	"context"

	"github.com/skylunna/agent-runtime/core"
)

// Transport 是服务端 <-> worker 的通信边界
// v1 实现: 同进程函数调用 (inprocess.go)
// v2 实现: gRPC 跨进程 (grpc.go)
// 上层代码只依赖这个接口，所以从 v1 到 v2，业务逻辑一行不改

// ---- Worker 视角：它需要的能力 ----
type WorkerTransport interface {
	// PollTask 向服务端领取一个待执行的步骤
	// v1 是直接函数调用；v2 是 gRPC long-poll
	PollTask(ctx context.Context) (*TaskAssignment, error)

	// ReportResult 把执行结果 (一批事件) 报告回服务端
	ReportResult(ctx context.Context, taskID string, events []core.Event) error
}

// ---- Server 视角：它需要的能力 ----
type ServerTransport interface {
	// SubmitTask 提交一个新任务 (来自 CLI 或 API)
	SubmitTask(ctx context.Context, task TaskSpec) (taskID string, err error)

	// Start 启动服务端的调度循环
	Start(ctx context.Context) error
}

// TaskAssignment: 服务端派给 worker 的一个工作单元
// 关键: 它带上了 "到目前为止的事件历史", 让 worker 能重放快进到当前状态
type TaskAssignment struct {
	TaskID  string
	Spec    TaskSpec
	History []core.Event // 靠重放这个恢复到崩溃前的状态
}

type TaskSpec struct {
	Goal  string   // agent 要完成的目标
	Tools []string // 允许使用的工具名
	Meta  map[string]string
}
