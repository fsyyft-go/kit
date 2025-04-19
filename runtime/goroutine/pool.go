// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"math"
	"time"

	"github.com/panjf2000/ants/v2"
)

// 默认配置值。
var (
	// 协程池的默认大小。
	sizeDefault = math.MaxInt32
	// 协程池中协程的默认过期时间。
	expiryDefault = time.Second
	// 是否默认预创建协程。
	preAllocDefault = false
	// 是否默认使用非阻塞模式。
	nonBlockingDefault = false
	// 默认的最大阻塞数量。
	maxBlockingDefault = 0
	// 默认的 panic 处理函数。
	panicHandlerDefault = func(r interface{}) {}
	// 是否默认提供指标信息。
	metricsDefault = true
)

// 类型定义。
type (
	// Option 定义了协程池的配置选项类型。
	Option func(p *goroutinePool)

	// GoroutinePool 定义了协程池的接口。
	// 该接口提供了协程池的基本操作，包括任务提交、容量调整和状态查询等功能。
	GoroutinePool interface {
		// Submit 提交一个任务到协程池中执行。
		// 参数：
		//   - task：要执行的任务函数。
		//
		// 返回值：
		//   - error：如果提交失败则返回错误。
		Submit(task func()) error

		// Tune 调整协程池的大小。
		// 参数：
		//   - size：新的协程池大小。
		Tune(size int)

		// Cap 获取协程池的容量大小。
		// 返回值：
		//   - int：协程池的容量。
		Cap() int

		// Running 获取协程池中正在运行的协程数量。
		// 返回值：
		//   - int：正在运行的协程数量。
		Running() int

		// Free 获取协程池中空闲的协程数量。
		// 返回值：
		//   - int：空闲的协程数量。
		Free() int

		// Waiting 获取协程池中等待执行的任务数量。
		// 返回值：
		//   - int：等待执行的任务数量。
		Waiting() int

		// IsClosed 检查协程池是否已经关闭。
		// 返回值：
		//   - bool：如果协程池已关闭则返回 true。
		IsClosed() bool
	}
)

// goroutinePool 实现了 GoroutinePool 接口，是协程池的具体实现。
type goroutinePool struct {
	// pool 是底层的 ants.Pool 实例，用于实际的任务调度和执行。
	pool *ants.Pool

	// 协程池大小（默认为 int 最大值）。
	size int
	// 协程池中协程的过期时间（默认为 1 秒）。
	expiry time.Duration
	// 是否在初始化协程池时预创建协程（默认为 false）。
	preAlloc bool
	// 是否非阻塞模式，非阻塞模式下添加任务时没有空闲协程会返回 err（默认为 false）。
	nonBlocking bool
	// 最大阻塞数量（默认为 0，表示不限制）。
	maxBlocking int
	// 子协程 panic 时回调方法（默认为空）。
	panicHandler func(interface{})

	// 协程池实例的名称，用于监控时区分不同实例（默认为空）。
	name string
	// 是否提供指标信息（默认为 true）。
	metrics bool

	// 用于通知子协程退出的通道。
	closed chan struct{}
}

// WithSize 设置协程池的大小。
// 参数：
//   - size：协程池的大小。
//
// 返回值：
//   - Option：配置选项函数。
func WithSize(size int) Option {
	return func(p *goroutinePool) {
		p.size = size
	}
}

// WithExpiry 设置协程池中协程的过期时间。
// 参数：
//   - expiry：协程的过期时间。
//
// 返回值：
//   - Option：配置选项函数。
func WithExpiry(expiry time.Duration) Option {
	return func(p *goroutinePool) {
		p.expiry = expiry
	}
}

// WithPreAlloc 设置是否在初始化时预创建协程。
// 参数：
//   - preAlloc：是否预创建协程。
//
// 返回值：
//   - Option：配置选项函数。
func WithPreAlloc(preAlloc bool) Option {
	return func(p *goroutinePool) {
		p.preAlloc = preAlloc
	}
}

// WithNonBlocking 设置是否使用非阻塞模式。
// 参数：
//   - nonBlocking：是否使用非阻塞模式。
//
// 返回值：
//   - Option：配置选项函数。
func WithNonBlocking(nonBlocking bool) Option {
	return func(p *goroutinePool) {
		p.nonBlocking = nonBlocking
	}
}

// WithMaxBlocking 设置最大阻塞数量。
// 参数：
//   - maxBlocking：最大阻塞数量。
//
// 返回值：
//   - Option：配置选项函数。
func WithMaxBlocking(maxBlocking int) Option {
	return func(p *goroutinePool) {
		p.maxBlocking = maxBlocking
	}
}

// WithPanicHandler 设置协程 panic 时的处理函数。
// 参数：
//   - panicHandler：panic 处理函数。
//
// 返回值：
//   - Option：配置选项函数。
func WithPanicHandler(panicHandler func(interface{})) Option {
	return func(p *goroutinePool) {
		p.panicHandler = panicHandler
	}
}

// WithName 设置协程池实例的名称。
// 参数：
//   - name：协程池实例的名称。
//
// 返回值：
//   - Option：配置选项函数。
func WithName(name string) Option {
	return func(p *goroutinePool) {
		p.name = name
	}
}

// WithMetrics 设置是否提供指标信息。
// 参数：
//   - metrics：是否提供指标信息。
//
// 返回值：
//   - Option：配置选项函数。
func WithMetrics(metrics bool) Option {
	return func(p *goroutinePool) {
		p.metrics = metrics
	}
}

// NewGoroutinePool 创建一个新的协程池实例。
// 参数：
//   - opts：配置选项。
//
// 返回值：
//   - GoroutinePool：新的协程池实例。
//   - func()：清理函数，用于释放协程池资源。
//   - error：如果创建失败则返回错误。
func NewGoroutinePool(opts ...Option) (GoroutinePool, func(), error) {
	// 创建协程池实例并设置默认值。
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

	// 应用用户提供的配置选项。
	for _, opt := range opts {
		opt(p)
	}

	// 定义清理函数，用于释放协程池资源。
	cleanup := func() {
		// 通知协程池关闭。
		p.closed <- struct{}{}
		// 如果底层池已创建，则释放资源。
		if p.pool != nil {
			errRelease := p.pool.ReleaseTimeout(10 * time.Second)
			if errRelease != nil {
				return
			}
		}
	}

	// 创建底层的 ants.Pool 实例。
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

	return p, cleanup, nil
}

// Submit 提交一个任务到协程池中执行。
// 参数：
//   - task：要执行的任务函数。
//
// 返回值：
//   - error：如果提交失败则返回错误。
func (p *goroutinePool) Submit(task func()) error {
	return p.pool.Submit(task)
}

// Tune 调整协程池的大小。
// 参数：
//   - size：新的协程池大小。
func (p *goroutinePool) Tune(size int) {
	p.pool.Tune(size)
}

// Cap 获取协程池的容量大小。
// 返回值：
//   - int：协程池的容量。
func (p *goroutinePool) Cap() int {
	return p.pool.Cap()
}

// Running 获取协程池中正在运行的协程数量。
// 返回值：
//   - int：正在运行的协程数量。
func (p *goroutinePool) Running() int {
	return p.pool.Running()
}

// Free 获取协程池中空闲的协程数量。
// 返回值：
//   - int：空闲的协程数量。
func (p *goroutinePool) Free() int {
	return p.pool.Free()
}

// Waiting 获取协程池中等待执行的任务数量。
// 返回值：
//   - int：等待执行的任务数量。
func (p *goroutinePool) Waiting() int {
	return p.pool.Waiting()
}

// IsClosed 检查协程池是否已经关闭。
// 返回值：
//   - bool：如果协程池已关闭则返回 true。
func (p *goroutinePool) IsClosed() bool {
	return p.pool.IsClosed()
}
