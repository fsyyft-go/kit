// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

//go:build !windows && arm64

package goroutine

func getg() *g

// Deprecated: GetGoID  获取 goroutine ID。
func GetGoID() int64 {
	return getg().goid
}
