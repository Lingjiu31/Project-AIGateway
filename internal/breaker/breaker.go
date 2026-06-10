package breaker

import (
	"sync"
	"time"

	"go.uber.org/zap"
)

type State int

const (
	StateClosed   State = iota // 正常，放行请求
	StateOpen                  // 熔断，直接拒绝
	StateHalfOpen              // 试探，放一批请求进来
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "healthy"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

type CircuitBreaker struct {
	mu          sync.Mutex
	name        string // 所属模型名，用于日志
	state       State
	failures    int           // 当前连续失败次数
	maxFailures int           // 触发熔断的阈值
	openAt      time.Time     // 进入 Open 状态的时间点
	timeout     time.Duration // Open 状态持续多久后转为 HalfOpen
}

func New(name string, maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:        name,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

// State 返回当前熔断器状态
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Allow 判断当前是否允许请求通过
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		// 正常状态，直接放行
		return true
	case StateOpen:
		// 熔断状态，先看等够时间没有
		if time.Since(cb.openAt) > cb.timeout {
			cb.state = StateHalfOpen
			zap.L().Info("熔断器进入半开状态，开始试探",
				zap.String("model", cb.name),
				zap.Duration("timeout", cb.timeout),
			)
			return true
		}
		return false
	default: // StateHalfOpen
		return true
	}
}

// ReportSuccess 上游请求成功，通知熔断器
func (cb *CircuitBreaker) ReportSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.failures = 0
		zap.L().Info("模型恢复正常，熔断器关闭",
			zap.String("model", cb.name),
		)
	}
}

// ReportFailure 上游请求失败，通知熔断器
func (cb *CircuitBreaker) ReportFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateHalfOpen:
		cb.state = StateOpen
		cb.openAt = time.Now()
		cb.failures = 0
		zap.L().Warn("试探失败，模型重新熔断",
			zap.String("model", cb.name),
			zap.Duration("retry_after", cb.timeout),
		)
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.openAt = time.Now()
			cb.failures = 0
			zap.L().Warn("模型熔断",
				zap.String("model", cb.name),
				zap.Int("failures", cb.maxFailures),
				zap.Duration("retry_after", cb.timeout),
			)
		}
	default:
	}
}
