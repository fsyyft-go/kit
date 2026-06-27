// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	// statTickTime 定义后台指标采集的时间间隔。
	statTickTime = 10 * time.Second
	// namespace 定义 Prometheus 指标命名空间。
	namespace = "kit_goroutine"
	// subsystem 定义 Prometheus 指标子系统名称。
	subsystem = "worker"
)

var (
	// MetricWorkerCurrent 记录协程池的当前状态。
	//
	// 标签：
	//   - name：协程池名称，对应 WithName 配置。
	//   - state：指标维度，可选值包括：
	//     - cap：协程池容量，对应 GoroutinePool.Cap。
	//     - running：正在执行任务的 worker 数量，对应 GoroutinePool.Running。
	//     - free：空闲 worker 数量，对应 GoroutinePool.Free。
	//     - waiting：等待调度的任务数量，对应 GoroutinePool.Waiting。
	MetricWorkerCurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "current",
		Help:      "goroutine pool's worker current.",
	}, []string{"name", "state"})
)

// stat 定期采集协程池运行状态并写入指标。
//
// 参数：
//   - p：需要持续采集状态的协程池实例。
func stat(p *goroutinePool) {
	ticker := time.NewTicker(statTickTime)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			collectWorkerMetrics(p)
		case <-p.closed:
			return
		}
	}
}

// collectWorkerMetrics 采集并写入协程池当前状态指标。
//
// 参数：
//   - p：需要写入指标的协程池实例。
func collectWorkerMetrics(p *goroutinePool) {
	MetricWorkerCurrent.WithLabelValues(p.name, "cap").Set(float64(p.pool.Cap()))
	MetricWorkerCurrent.WithLabelValues(p.name, "running").Set(float64(p.pool.Running()))
	MetricWorkerCurrent.WithLabelValues(p.name, "free").Set(float64(p.pool.Free()))
	MetricWorkerCurrent.WithLabelValues(p.name, "waiting").Set(float64(p.pool.Waiting()))
}
