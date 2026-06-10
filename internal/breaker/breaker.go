package breaker

import (
	"sync"
	"time"
)

type State int

const (
	StateClosed   State = iota // 正常，放行请求
	StateOpen                  // 熔断，直接拒绝
	StateHalfOpen              // 试探，放一批请求进来
)

type CircuitBreaker struct {
	mu          sync.Mutex
	state       State
	failures    int           // 当前连续失败次数
	maxFailures int           // 触发熔断的阈值
	openAt      time.Time     // 进入 Open 状态的时间点
	timeout     time.Duration // Open 状态持续多久后转为 HalfOpen
}

func New(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures: maxFailures,
		timeout:     timeout,
	}
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
			return true
		}
		return false
	default:
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
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
			cb.openAt = time.Now()
			cb.failures = 0
		}
	default:
	}
}
