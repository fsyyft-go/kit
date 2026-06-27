// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"math"
	"sync"
	"time"

	"github.com/panjf2000/ants/v2"

	kitlog "github.com/fsyyft-go/kit/log"
)

// 默认配置值。
var (
	// sizeDefault 定义默认协程池容量。
	// 该值使用 math.MaxInt32 近似表示默认情况下不主动收紧底层 ants.Pool 的容量上限。
	sizeDefault = math.MaxInt32
	// expiryDefault 定义空闲 worker 的默认回收周期。
	expiryDefault = time.Second
	// preAllocDefault 定义默认不预分配 worker。
	preAllocDefault = false
	// nonBlockingDefault 定义默认使用阻塞提交模式。
	nonBlockingDefault = false
	// maxBlockingDefault 定义阻塞提交模式下允许等待的最大任务数。
	maxBlockingDefault = 0
	// panicHandlerDefault 定义传给底层 ants.Pool 的默认 panic 回调。
	panicHandlerDefault = func(r interface{}) {}
	// metricsDefault 定义默认启用指标采集。
	metricsDefault = true

	// poolDefault 缓存包级默认协程池实例。
	poolDefault *goroutinePool
	// poolDefaultLocker 保护默认协程池的惰性初始化与后续读取。
	poolDefaultLocker sync.RWMutex
)

type (
	// Option 定义协程池配置修改函数。
	//
	// 参数：
	//   - p：待修改的协程池配置实例。
	Option func(p *goroutinePool)

	// GoroutinePool 定义协程池的任务提交、容量调整和状态查询能力。
	//
	// 该接口由 NewGoroutinePool 返回的实例实现。
	GoroutinePool interface {
		// Submit 提交一个任务到协程池中异步执行。
		//
		// 参数：
		//   - task：要交给协程池执行的任务函数。
		//
		// 返回：
		//   - error：底层协程池关闭或拒绝接收任务时返回错误。
		Submit(task func()) error

		// Tune 调整协程池的容量。
		//
		// 参数：
		//   - size：新的协程池容量。
		Tune(size int)

		// Cap 返回协程池当前容量。
		//
		// 参数：无。
		//
		// 返回：
		//   - int：协程池当前容量。
		Cap() int

		// Running 返回当前正在执行任务的 worker 数量。
		//
		// 参数：无。
		//
		// 返回：
		//   - int：当前正在执行任务的 worker 数量。
		Running() int

		// Free 返回当前空闲的 worker 数量。
		//
		// 参数：无。
		//
		// 返回：
		//   - int：当前空闲的 worker 数量。
		Free() int

		// Waiting 返回当前等待调度的任务数量。
		//
		// 参数：无。
		//
		// 返回：
		//   - int：当前等待调度的任务数量。
		Waiting() int

		// IsClosed 报告协程池是否已经关闭。
		//
		// 参数：无。
		//
		// 返回：
		//   - bool：协程池已关闭时返回 true。
		IsClosed() bool
	}
)

// goroutinePool 封装底层 ants.Pool，以及本包附加的指标采集和关闭信号。
type goroutinePool struct {
	// pool 是底层 ants 协程池实例，负责实际任务调度和 worker 管理。
	pool *ants.Pool

	// size 是传给 ants.NewPool 的容量配置。
	size int
	// expiry 是空闲 worker 的回收周期。
	expiry time.Duration
	// preAlloc 指示是否在创建时预分配 worker。
	preAlloc bool
	// nonBlocking 指示池满时是否立即返回，而不是阻塞等待空闲 worker。
	nonBlocking bool
	// maxBlocking 是阻塞提交模式下允许等待的最大任务数。
	maxBlocking int
	// panicHandler 是传给底层 ants.Pool 的 panic 回调。
	panicHandler func(interface{})

	// name 用于区分不同协程池实例的指标标签。
	name string
	// metrics 指示是否启动后台指标采集协程。
	metrics bool

	// closed 用于通知指标采集协程退出。
	closed chan struct{}
}

// WithSize 设置协程池容量。
//
// 参数：
//   - size：传给底层 ants.NewPool 的容量值。
//
// 返回：
//   - Option：用于更新协程池容量配置的选项函数。
func WithSize(size int) Option {
	return func(p *goroutinePool) {
		p.size = size
	}
}

// WithExpiry 设置空闲 worker 的回收周期。
//
// 参数：
//   - expiry：空闲 worker 在底层池中保留的最长空闲时间。
//
// 返回：
//   - Option：用于更新 worker 回收周期的选项函数。
func WithExpiry(expiry time.Duration) Option {
	return func(p *goroutinePool) {
		p.expiry = expiry
	}
}

// WithPreAlloc 设置是否在初始化时预分配 worker。
//
// 参数：
//   - preAlloc：为 true 时在创建协程池时预分配 worker。
//
// 返回：
//   - Option：用于更新预分配策略的选项函数。
func WithPreAlloc(preAlloc bool) Option {
	return func(p *goroutinePool) {
		p.preAlloc = preAlloc
	}
}

// WithNonBlocking 设置是否启用非阻塞提交模式。
//
// 参数：
//   - nonBlocking：为 true 时在池满后立即返回提交错误，而不是等待空闲 worker。
//
// 返回：
//   - Option：用于更新提交阻塞策略的选项函数。
func WithNonBlocking(nonBlocking bool) Option {
	return func(p *goroutinePool) {
		p.nonBlocking = nonBlocking
	}
}

// WithMaxBlocking 设置阻塞提交模式下允许等待的最大任务数。
//
// 参数：
//   - maxBlocking：传给 ants.WithMaxBlockingTasks 的最大等待任务数。
//
// 返回：
//   - Option：用于更新最大阻塞任务数的选项函数。
func WithMaxBlocking(maxBlocking int) Option {
	return func(p *goroutinePool) {
		p.maxBlocking = maxBlocking
	}
}

// WithPanicHandler 设置底层协程池的 panic 回调。
//
// 参数：
//   - panicHandler：传给 ants.WithPanicHandler 的回调函数，用于处理 worker 执行任务时发生的 panic。
//
// 返回：
//   - Option：用于更新 panic 回调的选项函数。
func WithPanicHandler(panicHandler func(interface{})) Option {
	return func(p *goroutinePool) {
		p.panicHandler = panicHandler
	}
}

// WithName 设置协程池实例名称。
//
// 参数：
//   - name：写入指标标签的协程池名称；为空时指标使用空名称标签。
//
// 返回：
//   - Option：用于更新协程池名称的选项函数。
func WithName(name string) Option {
	return func(p *goroutinePool) {
		p.name = name
	}
}

// WithMetrics 设置是否启用指标采集。
//
// 参数：
//   - metrics：为 true 时启动后台采集协程并持续更新 MetricWorkerCurrent。
//
// 返回：
//   - Option：用于更新指标采集开关的选项函数。
func WithMetrics(metrics bool) Option {
	return func(p *goroutinePool) {
		p.metrics = metrics
	}
}

// NewGoroutinePool 创建一个新的协程池实例。
//
// 参数：
//   - opts：可选配置项，按传入顺序覆盖默认配置。
//
// 返回：
//   - GoroutinePool：创建成功的协程池实例。
//   - func()：cleanup 函数；调用方在不再使用该实例时应调用一次，用于停止指标采集协程并释放底层 ants.Pool 资源。
//   - error：底层 ants.NewPool 创建失败时返回错误；此时协程池实例和 cleanup 都为 nil。
func NewGoroutinePool(opts ...Option) (GoroutinePool, func(), error) {
	p := &goroutinePool{
		size:         sizeDefault,
		expiry:       expiryDefault,
		preAlloc:     preAllocDefault,
		nonBlocking:  nonBlockingDefault,
		maxBlocking:  maxBlockingDefault,
		panicHandler: panicHandlerDefault,
		metrics:      metricsDefault,
		closed:       make(chan struct{}, 1),
	}

	for _, opt := range opts {
		opt(p)
	}

	cleanup := func() {
		// cleanup 设计为单次调用；closed 是退出信号通道，不是可重复广播的 close。
		p.closed <- struct{}{}
		if p.pool != nil {
			errRelease := p.pool.ReleaseTimeout(10 * time.Second)
			if errRelease != nil {
				return
			}
		}
	}

	pool, errNewPool := ants.NewPool(
		p.size,
		ants.WithExpiryDuration(p.expiry),
		ants.WithPreAlloc(p.preAlloc),
		ants.WithNonblocking(p.nonBlocking),
		ants.WithMaxBlockingTasks(p.maxBlocking),
		ants.WithPanicHandler(p.panicHandler),
	)
	if errNewPool != nil {
		return nil, nil, errNewPool
	}
	p.pool = pool

	if p.metrics {
		go stat(p)
	}

	return p, cleanup, nil
}

// Submit 提交一个任务到协程池中执行。
//
// 参数：
//   - task：要交给底层 ants.Pool 执行的任务函数。
//
// 返回：
//   - error：底层协程池关闭或拒绝接收任务时返回错误。
func (p *goroutinePool) Submit(task func()) error {
	return p.pool.Submit(task)
}

// Tune 调整协程池的容量。
//
// 参数：
//   - size：新的协程池容量。
func (p *goroutinePool) Tune(size int) {
	p.pool.Tune(size)
}

// Cap 返回协程池当前容量。
//
// 参数：无。
//
// 返回：
//   - int：协程池当前容量。
func (p *goroutinePool) Cap() int {
	return p.pool.Cap()
}

// Running 返回当前正在执行任务的 worker 数量。
//
// 参数：无。
//
// 返回：
//   - int：当前正在执行任务的 worker 数量。
func (p *goroutinePool) Running() int {
	return p.pool.Running()
}

// Free 返回当前空闲的 worker 数量。
//
// 参数：无。
//
// 返回：
//   - int：当前空闲的 worker 数量。
func (p *goroutinePool) Free() int {
	return p.pool.Free()
}

// Waiting 返回当前等待调度的任务数量。
//
// 参数：无。
//
// 返回：
//   - int：当前等待调度的任务数量。
func (p *goroutinePool) Waiting() int {
	return p.pool.Waiting()
}

// IsClosed 报告协程池是否已经关闭。
//
// 参数：无。
//
// 返回：
//   - bool：协程池已关闭时返回 true。
func (p *goroutinePool) IsClosed() bool {
	return p.pool.IsClosed()
}

// Submit 将 task 提交到包级默认协程池执行。
//
// 首次调用会惰性创建默认池。与显式 GoroutinePool.Submit 不同，包级包装层会 recover task panic 并记录日志，
// 不会把 panic 继续向调用方传播，也不会通过返回值暴露该 panic。
//
// 参数：
//   - task：要提交到包级默认协程池执行的任务函数。
//
// 返回：
//   - error：默认池初始化失败或底层提交失败时返回错误。task panic 会被 recover 并记录日志，不通过返回值暴露。
func Submit(task func()) error {
	p, err := defaultPool()
	if err != nil {
		return err
	}

	return p.Submit(func() {
		defer func() {
			if r := recover(); nil != r {
				kitlog.Error("goroutine panic", r)
			}
		}()
		task()
	})
}

// defaultPool 获取默认协程池实例，并在首次调用时完成惰性初始化。
//
// 该函数使用 poolDefaultLocker 串行化 poolDefault 的创建与读取，避免并发 Submit 时出现数据竞争。
//
// 参数：无。
//
// 返回：
//   - *goroutinePool：默认协程池实例。
//   - error：默认协程池初始化失败时返回错误。
func defaultPool() (*goroutinePool, error) {
	poolDefaultLocker.Lock()
	defer poolDefaultLocker.Unlock()

	if poolDefault != nil {
		return poolDefault, nil
	}

	p, cleanup, err := NewGoroutinePool(WithName("default"))
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, err
	}

	poolDefault = p.(*goroutinePool)
	return poolDefault, nil
}
