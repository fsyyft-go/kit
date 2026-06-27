// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package goroutine

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectWorkerMetrics_UpdatesWorkerMetrics 验证单次指标采样会写入 worker 状态。
//
// 该测试使用唯一指标 label 隔离全局 Prometheus 状态，并直接调用采样函数覆盖 cap、running、free、waiting
// 四类指标写入行为，避免单元测试绑定生产 10 秒定时器。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestCollectWorkerMetrics_UpdatesWorkerMetrics(t *testing.T) {
	const givePoolName = "test-collect-worker-metrics-updates-worker-metrics"

	deleteWorkerMetricsForTest(givePoolName)
	t.Cleanup(func() {
		deleteWorkerMetricsForTest(givePoolName)
	})

	pool := newGoroutinePoolForTest(t,
		WithSize(3),
		WithName(givePoolName),
		WithMetrics(false),
	)
	concretePool, ok := pool.(*goroutinePool)
	require.True(t, ok)

	// 验证单次采样会将当前协程池容量、运行数、空闲数和等待数写入指标。
	collectWorkerMetrics(concretePool)

	assert.Equal(t, float64(pool.Cap()), workerMetricValue(t, givePoolName, "cap"))
	assert.Equal(t, float64(pool.Running()), workerMetricValue(t, givePoolName, "running"))
	assert.Equal(t, float64(pool.Free()), workerMetricValue(t, givePoolName, "free"))
	assert.Equal(t, float64(pool.Waiting()), workerMetricValue(t, givePoolName, "waiting"))
}

// TestStat_StopsWhenPoolIsClosed 验证指标采集协程在协程池关闭信号到达时退出。
//
// 该测试直接发送关闭信号并等待 stat 返回，覆盖定时采集循环的退出分支，且不依赖生产 ticker 周期。
//
// 参数：
//   - t: 测试上下文，用于运行子测试和报告断言失败。
func TestStat_StopsWhenPoolIsClosed(t *testing.T) {
	pool := newGoroutinePoolForTest(t,
		WithSize(1),
		WithName("test-stat-stops-when-pool-is-closed"),
		WithMetrics(false),
	)
	concretePool, ok := pool.(*goroutinePool)
	require.True(t, ok)

	stopped := make(chan struct{})
	go func() {
		stat(concretePool)
		close(stopped)
	}()

	// 验证 stat 收到关闭信号后能够及时退出，避免指标采集 goroutine 泄漏。
	concretePool.closed <- struct{}{}
	receiveWithin(t, stopped, "metrics collector shutdown")
}

// deleteWorkerMetricsForTest 删除指定协程池名称的 worker 指标。
//
// 该辅助函数用于隔离全局 Prometheus GaugeVec 状态，避免重复运行测试时读取到旧 label 值。
//
// 参数：
//   - name: 需要删除指标的协程池名称 label。
func deleteWorkerMetricsForTest(name string) {
	for _, state := range []string{"cap", "running", "free", "waiting"} {
		MetricWorkerCurrent.DeleteLabelValues(name, state)
	}
}

// workerMetricValue 读取指定 worker 指标当前值。
//
// 该辅助函数封装 Prometheus Gauge 写出逻辑，使测试断言聚焦于协程池指标语义。
//
// 参数：
//   - t: 测试上下文，用于报告指标读取失败并标记辅助函数调用栈。
//   - name: 协程池名称 label。
//   - state: 协程池状态 label。
//
// 返回：
//   - float64: 指定 label 组合对应的 Gauge 当前值。
func workerMetricValue(t *testing.T, name string, state string) float64 {
	t.Helper()

	metric := &dto.Metric{}
	require.NoError(t, MetricWorkerCurrent.WithLabelValues(name, state).Write(metric))
	require.NotNil(t, metric.Gauge)
	return metric.Gauge.GetValue()
}
