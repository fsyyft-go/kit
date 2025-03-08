// Copyright 2024 The fsyyft Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/fsyyft-go/kit/config"
)

func main() {
	fmt.Printf("Version: %s\n", config.CurrentVersion.Version())
	fmt.Printf("Git Version: %s\n", config.CurrentVersion.GitVersion())
	fmt.Printf("Build Time: %s\n", config.CurrentVersion.BuildTimeString())
	fmt.Printf("Library Directory: %s\n", config.CurrentVersion.BuildLibraryDirectory())
	fmt.Printf("Working Directory: %s\n", config.CurrentVersion.BuildWorkingDirectory())
	fmt.Printf("GOPATH Directory: %s\n", config.CurrentVersion.BuildGopathDirectory())
	fmt.Printf("GOROOT Directory: %s\n", config.CurrentVersion.BuildGorootDirectory())
	fmt.Printf("Debug Mode: %v\n", config.CurrentVersion.Debug())
}
