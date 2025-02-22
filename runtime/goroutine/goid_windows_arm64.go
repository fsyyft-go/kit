// Copyright (c) 2025 fsyyft-go
//
// Licensed under the MIT License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://github.com/fsyyft-go/kit/blob/main/LICENSE
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build linux && arm64

package goroutine

// Deprecated: GetGoID  获取 goroutine ID。
func GetGoID() int64 {
	// TODO 汇编的方法未实现，先使用开销较大的。
	return getGoIDSlow()
}

// GetGoIDSlow 获取当前协程的 ID，当无法从 GetGoID 获取协程 ID 时使用此方法。
// 该方法通过获取协程的堆栈信息，然后解析堆栈信息来提取协程 ID。
func GetGoIDSlow() int64 {
	return getGoIDSlow()
}
