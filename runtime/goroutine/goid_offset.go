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

//go:build amd64

package goroutine

import (
	"runtime"
	"strings"
)

var (
	offsetDict = map[string]int64{
		"go1.4":  128,
		"go1.5":  184,
		"go1.6":  192,
		"go1.7":  192,
		"go1.8":  192,
		"go1.9":  152,
		"go1.10": 152,
		"go1.11": 152,
		"go1.12": 152,
		"go1.13": 152,
		"go1.14": 152,
		"go1.15": 152,
		"go1.16": 152,
		"go1.17": 152,
		"go1.18": 152,
		"go1.19": 152,
		"go1.20": 152,
		"go1.21": 152,
		"go1.22": 152,
		"go1.23": 160,
	}

	offset = func() int64 {
		ver := strings.Join(strings.Split(runtime.Version(), ".")[:2], ".")
		return offsetDict[ver]
	}()
)

func Offset() int64 {
	return offset
}
