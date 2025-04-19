// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// 定义协程池指标统计相关的常量。
const (
	// statTickTime 定义指标采集的时间间隔，默认为 10 秒。
	statTickTime = 10 * time.Second
	// namespace 定义 prometheus 指标的命名空间。
	namespace = "kit_goroutine"
	// subsystem 定义 prometheus 指标的子系统名称。
	subsystem = "worker"
)

var (
	// MetricWorkerCurrent 用于记录协程池的当前状态指标。
	// 该指标包含以下标签：
	// - name: 协程池的名称。
	// - state: 协程池的状态，包括容量、运行中、空闲和等待中的协程数量。
	MetricWorkerCurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "current",
		Help:      "goroutine pool's worker current.",
	}, []string{"name", "state"})
)

// stat 定期采集协程池的运行状态指标。
// 该函数会启动一个定时器，每 10 秒采集一次协程池的状态信息。
// 采集的指标包括：
// - 协程池的总容量。
// - 当前正在运行的协程数量。
// - 当前空闲的协程数量。
// - 当前等待任务的协程数量。
// 当协程池关闭时，该函数会自动退出。
func stat(p *goroutinePool) {
	// 创建定时器，每 10 秒触发一次。
	ticker := time.NewTicker(statTickTime)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 更新协程池的容量指标。
			MetricWorkerCurrent.WithLabelValues(p.name, "cap").Set(float64(p.pool.Cap()))
			// 更新正在运行的协程数量指标。
			MetricWorkerCurrent.WithLabelValues(p.name, "running").Set(float64(p.pool.Running()))
			// 更新空闲协程数量指标。
			MetricWorkerCurrent.WithLabelValues(p.name, "free").Set(float64(p.pool.Free()))
			// 更新等待任务的协程数量指标。
			MetricWorkerCurrent.WithLabelValues(p.name, "waiting").Set(float64(p.pool.Waiting()))
		case <-p.closed:
			// 当协程池关闭时退出循环。
			return
		}
	}
}
