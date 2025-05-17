// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package pinyin

import (
	"testing"
)

func TestPinyin(t *testing.T) {
	p := NewPinyin(WithStyle(Tone))
	t.Log(p.Pinyin("中国"))
}
