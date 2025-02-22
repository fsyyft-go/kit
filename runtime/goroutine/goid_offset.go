// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

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
		"go1.24": 160,
	}

	offset = func() int64 {
		ver := strings.Join(strings.Split(runtime.Version(), ".")[:2], ".")
		return offsetDict[ver]
	}()
)

func Offset() int64 {
	return offset
}
