// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package context

import (
	"context"
	"testing"
	"time"
)

func TestWithoutCancel(t *testing.T) {
	ch := make(chan struct{}, 3)
	f := func(ctx context.Context, name string) {
		timer := time.NewTimer(time.Second)
		select {
		case <-ctx.Done():
			t.Logf("goroutine %s 强制退出信号", name)
		case <-timer.C:
			t.Logf("goroutine %s 完成任务信号", name)
		}
		ch <- struct{}{}
	}
	ctx, cancel := context.WithCancelCause(context.Background())
	ctxTo, _ := context.WithTimeout(ctx, 10*time.Millisecond)
	t.Run("超时停止", func(t *testing.T) {

		go f(ctxTo, t.Name())
	})
	t.Run("正常停止", func(t *testing.T) {
		go f(ctx, t.Name())
	})
	t.Run("阻断停止", func(t *testing.T) {
		go f(context.WithoutCancel(ctxTo), t.Name())
	})
	time.Sleep(50 * time.Millisecond)
	cancel(context.Canceled)
	<-ch
	<-ch
	<-ch
}
