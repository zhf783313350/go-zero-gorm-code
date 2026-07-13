// Package event 提供轻量级的应用内异步事件总线。
//
// 设计模式：Outbox / 观察者模式（无需外部 MQ 依赖）
//
// 架构说明：
//   - EventBus 持有一个有缓冲的 channel，作为内存消息队列（Outbox）。
//   - 业务层（Logic）调用 Publish 投递事件，立即返回，不阻塞主流程。
//   - main.go 启动时通过 Start() 开启后台 goroutine 消费事件（审计日志、通知等）。
//   - 若未来需要接入 Kafka/RabbitMQ，只需将 consumer goroutine 内的处理替换为
//     MQ 的 Producer.Send 调用即可，业务层代码完全无需变更（开闭原则）。
package event

import (
	"github.com/zeromicro/go-zero/core/logx"
)

// EventType 事件类型枚举
type EventType string

const (
	EventUserCreated EventType = "user.created"
	EventUserUpdated EventType = "user.updated"
	EventUserDeleted EventType = "user.deleted"
	EventUserLogin   EventType = "user.login"
)

// UserEvent 用户领域事件结构体
type UserEvent struct {
	Type        EventType `json:"type"`
	PhoneNumber string    `json:"phoneNumber"`
	UserID      int64     `json:"userId,omitempty"`
	Operator    string    `json:"operator,omitempty"` // 预留：操作人信息
}

// EventBus 内存异步事件总线（可平滑替换为 Kafka Producer）
type EventBus struct {
	ch chan UserEvent
}

// NewEventBus 创建事件总线，bufferSize 为缓冲队列容量。
// 建议生产环境设置为 1000-5000，防止业务高峰期突发事件堆积。
func NewEventBus(bufferSize int) *EventBus {
	return &EventBus{
		ch: make(chan UserEvent, bufferSize),
	}
}

// Publish 异步投递事件（非阻塞）。
// 若 channel 满，会记录警告日志并丢弃该事件（可根据需要改为阻塞或持久化兜底）。
func (b *EventBus) Publish(evt UserEvent) {
	select {
	case b.ch <- evt:
	default:
		logx.Errorf("[EventBus] channel is full, dropping event: %v", evt.Type)
	}
}

// Start 启动后台消费 goroutine，通过 stopCh 接收优雅停机信号。
// 应在 main.go 中调用，并在 server.Stop() 前关闭 stopCh。
func (b *EventBus) Start(stopCh <-chan struct{}) {
	go func() {
		logx.Info("[EventBus] consumer goroutine started")
		for {
			select {
			case evt := <-b.ch:
				b.consume(evt)
			case <-stopCh:
				// 优雅停机：排空剩余事件后退出
				for {
					select {
					case evt := <-b.ch:
						b.consume(evt)
					default:
						logx.Info("[EventBus] consumer goroutine stopped gracefully")
						return
					}
				}
			}
		}
	}()
}

// consume 事件消费处理器。
// 当前实现：结构化审计日志。
// 扩展点：可在此处替换为 Kafka/RabbitMQ Producer.Send，实现真正的消息队列投递。
func (b *EventBus) consume(evt UserEvent) {
	switch evt.Type {
	case EventUserCreated:
		logx.Infof("[AuditLog] USER_CREATED | phone=%s", evt.PhoneNumber)
	case EventUserUpdated:
		logx.Infof("[AuditLog] USER_UPDATED | phone=%s", evt.PhoneNumber)
	case EventUserDeleted:
		logx.Infof("[AuditLog] USER_DELETED | phone=%s | userId=%d", evt.PhoneNumber, evt.UserID)
	case EventUserLogin:
		logx.Infof("[AuditLog] USER_LOGIN   | phone=%s | userId=%d", evt.PhoneNumber, evt.UserID)
	default:
		logx.Infof("[AuditLog] UNKNOWN_EVENT | type=%s", evt.Type)
	}
}
