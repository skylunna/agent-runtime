package ports

import (
	"context"

	"github.com/skylunna/agent-runtime/core"
)

// EventStore 是事件的唯一归宿
// 整个系统只通过这个接口读写事件，绝不直接碰文件或数据库
type EventStore interface {
	// Append 追加一批事件。注意是批量 - 一个步骤可能产生多个事件
	// 它们要么全部落盘，要么全部失败 （原子性）
	// expectedSeq 用于乐观并发空值：如果当前最大 Seq 不等于它，说明有并发冲突
	Append(ctx context.Context, taskID string, expectedSeq int64, events []core.Event) error

	// Load 读取一个任务的全部事件，按 Seq 升序。重放就靠它
	Load(ctx context.Context, taskID string) ([]core.Event, error)

	// LoadFrom 从某个 Seq 之后读取 (增量重放 / 快进时用)
	LoadFrom(ctx context.Context, taskID string, afterSeq int64) ([]core.Event, error)
}
